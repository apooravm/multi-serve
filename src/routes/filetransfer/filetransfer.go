package filetransfer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	// "strconv"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// Different initial byte types different
// 1 -> Register a Sender
// 2 -> Server responds with unique_code
// 3 -> Register a Receiver
// 4 -> Server responds to Sender to begin transfer
// 5 -> Transfer packet from Sender to Receiver
// 6 -> TBD
// Something for if either client disconnects.
const (
	// Register a sender
	InitialTypeRegisterSender = uint8(0x01)

	// Server responds back to sender with a unique code
	InitialTypeUniqueCode = uint8(0x02)

	// Register a receiver
	InitialTypeRegisterReceiver = uint8(0x03)

	// Server sends metadata of the transfer to the receiver
	InitialTypeTransferMetaData = uint8(0x04)

	// Server responds back to sender to begin transfer
	// Receiver responds with 1 or 0
	// 1 to begin transfer. 0 to abort.
	InitialTypeBeginTransfer = uint8(0x05)

	// Transfer packet from sender to receiver.
	InitialTypeTransferPacket = uint8(0x06)

	// Volutanry disconnection
	InitialTypeCloseConn = uint8(0x08)

	// Text message about some issue or error or whatever
	InitialTypeTextMessage = uint8(0x09)

	// current version
	version = byte(1)
)

// Been handling this the wrong way all this while.
// Since the ws connection persists, i just need to update the global newFTMeta which has all the info
// Dont need to send unique_code with every request
var (
	newFTMeta FTMeta
)

// Since handshake is a 1 time thing, it will be done through json
type ClientHandshake struct {
	Version uint8
	// Send => 0, Receive => 1
	Intent     uint8
	UniqueCode uint8
	FileSize   uint64
	ClientName string
	Filename   string
}

// Metadata for receiver from server
type MDReceiver struct {
	FileSize   uint64
	SenderName string
	Filename   string
}

func FileTransferWs(c echo.Context) error {
	ConnUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := ConnUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("WS_ERR:", err.Error())
		return err
	}

	defer conn.Close()

	// client_id := strconv.Itoa(utils.Id_Gen.GenerateNewID())
	for {
		// Sudden disconnect makes this throw err
		// Incase either suddenly disconnect
		// Send a message to the other and close the connection.
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Could not read", err.Error())
			_ = conn.Close()
			return nil
		}

		if messageType != websocket.BinaryMessage {
			utils.LogData("E:FT unexpected message type")
		}

		if len(message) < 1 {
			fmt.Println("Empty payload, disconnecting conn.")
			conn.Close()
			return nil
		}

		// Different initial types
		switch message[0] {
		case InitialTypeRegisterSender:
			HandleRegisterSender(message, conn)

		case InitialTypeRegisterReceiver:
			HandleRegisterReceiver(message, conn)

		case InitialTypeBeginTransfer:
			if len(message) != 2 {
				newMsg := []byte("Invalid frame. Need [initial_byte][trigger_byte]")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(conn, newMsg)
				_ = conn.Close()
				return nil
			}

			var beginTransferOrNo uint8 = message[1]

			// Receiver somehow moves to this step without registering
			// Should not happen, but just in case.
			if newFTMeta.ReceiverConn == nil {
				newMsg := []byte("Receiver not registered. Disconnecting.")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(conn, newMsg)
				_ = conn.Close()
				return nil
			}

			beginTransferRes := new(bytes.Buffer)

			if err := binary.Write(beginTransferRes, binary.BigEndian, InitialTypeBeginTransfer); err != nil {
				fmt.Println("E:Creating response. idk man fk this is getting annoying", err.Error())
				_ = conn.Close()
				return nil
			}

			// Write the same uint8 to the packet being sent to the sender
			if err := binary.Write(beginTransferRes, binary.BigEndian, beginTransferOrNo); err != nil {
				fmt.Println("E:Creating response. idk man fk this is getting annoying", err.Error())
				_ = conn.Close()
				return nil
			}

			// Really gotta do smn about these funcs
			// Getting annoying writing them again and again
			if err := newFTMeta.SenderConn.WriteMessage(websocket.BinaryMessage, beginTransferRes.Bytes()); err != nil {
				fmt.Println("E:Writing message to conn", err.Error())
				_ = conn.Close()
				return nil
			}

		case InitialTypeTransferPacket:
			if err := newFTMeta.ReceiverConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				fmt.Println("Couldnt forward packet to receiver")
				return nil
			}

			// Packet frame
		// 0 -> sender, 1 -> receiver
		// [init_byte][unique_code][sender_or_reic]
		// May not need this one here
		case InitialTypeCloseConn:
			// Identify by conns by the unique_code.
			if len(message) != 3 {
				newMsg := []byte("Disconnecting but cannot diconnect the receiver without the code.")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(conn, newMsg)
				_ = conn.Close()
				return nil
			}

			var unique_code uint8 = message[1]
			var isSenderOrReceiver uint8 = message[2]

			ongoingFT, exists := FTMap.GetClient(string(unique_code))
			if !exists {
				newMsg := []byte("Disconnecting.")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(conn, newMsg)
				_ = conn.Close()
				return nil
			}

			// Sender has disconnected
			if isSenderOrReceiver == 0 {

				newMsg := []byte("Disconnecting")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(ongoingFT.SenderConn, newMsg)
				_ = ongoingFT.SenderConn.Close()

				newMsg = []byte("Sender left. Disconnecting")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(ongoingFT.ReceiverConn, newMsg)
				_ = ongoingFT.ReceiverConn.Close()

				return nil
			}

			// Receiver has disconnected
			if isSenderOrReceiver == 1 {
				newMsg := []byte("Disconnecting")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(ongoingFT.SenderConn, newMsg)
				_ = ongoingFT.SenderConn.Close()

				newMsg = []byte("Receiver left. Disconnecting")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(ongoingFT.SenderConn, newMsg)
				_ = ongoingFT.SenderConn.Close()
			}

			FTMap.DeleteClient(string(unique_code))
			return nil

		default:
			newMsg := []byte("Bro what? Unknown initial type.")
			newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
			_ = ConnWriteMessage(conn, newMsg)
			_ = conn.Close()
			return nil
		}

	}
}

func ConnWriteMessage(conn *websocket.Conn, message []byte) error {
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		utils.LogData("E:Writing to websocket. Message was", string(message))
		return err
	}

	return nil
}

func HandleRegisterSender(message []byte, conn *websocket.Conn) error {
	var clientHandshake ClientHandshake
	// Ignore the initial byte
	if err := json.Unmarshal(message[1:], &clientHandshake); err != nil {
		fmt.Println("Could not unmarshal", err.Error())
		_ = conn.Close()
		return nil
	}

	// If clientHandshake.UniqueCode == 0 -> register sender
	// Else -> already registered. tf is it doing here again?

	// if clientHandshake.UniqueCode != 0 {
	// 	fmt.Println("Already registered. Waiting for receiver but this shouldnt have been sent again. Closing")
	// 	conn.Close()
	// 	return nil
	// }

	// This could be done better.
	// Only 255 unique ones possible
	unique_code := FTCodeGenerator.NewCode()
	responseBfr := new(bytes.Buffer)

	if err := binary.Write(responseBfr, binary.BigEndian, InitialTypeUniqueCode); err != nil {
		fmt.Println("Could not create response bruh", err.Error())
		_ = conn.Close()
		return nil
	}

	if err := binary.Write(responseBfr, binary.BigEndian, unique_code); err != nil {
		fmt.Println("Could not create response", err.Error())
		_ = conn.Close()
		return nil
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, responseBfr.Bytes()); err != nil {
		utils.LogData("E:Could not send unique code")
		_ = conn.Close()
		return nil
	}

	// Says client but actually is the whole process
	newFTMeta.Version = version
	newFTMeta.Code = unique_code
	newFTMeta.SenderConn = conn
	newFTMeta.SenderName = clientHandshake.ClientName
	newFTMeta.FileSize = clientHandshake.FileSize
	newFTMeta.Filename = clientHandshake.Filename

	FTMap.AddClient(string(unique_code), &newFTMeta)

	return nil
}

func HandleRegisterReceiver(message []byte, conn *websocket.Conn) error {
	// Receiver packet
	// [initial_byte][unique_code][receiver_name]
	incomingBuffer := bytes.NewReader(message[1:])
	var incomingReceiverCode uint8
	// var incomingReceiverName string

	if err := binary.Read(incomingBuffer, binary.BigEndian, &incomingReceiverCode); err != nil {
		fmt.Println("Could not read incoming receiver register unique code", err.Error())
		_ = conn.Close()
		return nil
	}

	// the read position in the buffer changes with every read
	// this just reads everything from the current position till the end
	incomingReceiverNameBytes, err := io.ReadAll(incomingBuffer)
	if err != nil {
		fmt.Println("Could not read incoming receiver register name", err.Error())
		_ = conn.Close()
		return nil
	}

	// if err := binary.Read(incomingBuffer, binary.BigEndian, incomingReceiverName); err != nil {
	// 	fmt.Println("Could not read incoming receiver register name", err.Error())
	// 	_ = conn.Close()
	// 	return nil
	// }

	incomingReceiverName := string(incomingReceiverNameBytes)

	ongoing, exists := FTMap.GetClient(string(incomingReceiverCode))
	if !exists {
		newMsg := []byte("Sender not found.")
		newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
		_ = ConnWriteMessage(conn, newMsg)
		_ = conn.Close()
		return nil
	}

	FTMap.UpdateClient(string(incomingReceiverCode), func(client FTMeta) FTMeta {
		client.ReceiverConn = conn
		client.ReceiverName = incomingReceiverName
		return client
	})

	// Update the newFTMeta for the current client
	ongoingFT, exists := FTMap.GetClient(string(incomingReceiverCode))
	if exists {
		newFTMeta = ongoingFT

	}

	// After the FTMap has been updated with the receiver's data

	// TransferMD packet frame
	// [initial_byte][json_transferMD]
	// transferMD := new(bytes.Buffer)
	//
	// if err := binary.Write(transferMD, binary.BigEndian, InitialTypeTransferMetaData); err != nil {
	// 	fmt.Println("Could not create binary response", err.Error())
	// 	_ = conn.Close()
	// 	return nil
	// }

	transferMD := MDReceiver{
		FileSize:   ongoing.FileSize,
		SenderName: ongoing.SenderName,
		Filename:   ongoing.Filename,
	}

	resByteArr, err := json.Marshal(&transferMD)
	if err != nil {
		fmt.Println("Could not marshal metadata response", err.Error())
		_ = conn.Close()
		return nil
	}

	resByteArr = append([]byte{InitialTypeTransferMetaData}, resByteArr...)
	if err := conn.WriteMessage(websocket.BinaryMessage, resByteArr); err != nil {
		fmt.Println("Could not write to receiver", err.Error())
		_ = conn.Close()
		return nil
	}

	return nil
}

// func broadcastToSubscribers(data string) {
// 	for _, client := range SubscribedUsersMap.Clients {
// 		if err := client.Conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
// 			utils.LogDataToPath(utils.DUMMY_WS_LOG_PATH, "BROAD_ALL_ERR:", err.Error())
// 			return
// 		}
// 	}
// }

// Handle Client socket disconnection
// Graceful handling prevents error logs
func clientDisconnect(conn *websocket.Conn, clientID string) {
	// if _, isSubbed := SubscribedUsersMap.GetClient(clientID); isSubbed {
	// 	SubscribedUsersMap.DeleteClient(clientID)
	// }
	// if _, isFound := UserMap.GetClient(clientID); isFound {
	// 	UserMap.DeleteClient(clientID)
	// }

	conn.Close()
	// Handle disconnection or error here
	// // Delete client from the map
	// ChatClientsMap.DeleteClient(client_id)
	// message := client.Username + " left!"
	// LogData(message, CHAT_LOG)
	// BroadcastServerMessageAll(message)
}
