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

	// Text message about some issue or error or whatever
	InitialTypeTextMessage = uint8(0x09)

	// current version
	version = byte(1)
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
			var newFTMeta FTMeta
			newFTMeta.Version = version
			newFTMeta.Code = unique_code
			newFTMeta.SenderConn = conn
			newFTMeta.SenderName = clientHandshake.ClientName
			newFTMeta.FileSize = clientHandshake.FileSize
			newFTMeta.Filename = clientHandshake.Filename

			FTMap.AddClient(string(unique_code), &newFTMeta)

		case InitialTypeRegisterReceiver:
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

		case InitialTypeBeginTransfer:
			if len(message) != 3 {
				newMsg := []byte("Invalid frame. Need [initial_byte][trigger_byte][unique_code(1byte)]")
				newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
				_ = ConnWriteMessage(conn, newMsg)
				_ = conn.Close()
				return nil
			}

			var beginTransferOrNo uint8 = message[1]

			if beginTransferOrNo == 0 {
				// Abort transfer

			} else if beginTransferOrNo == 1 {
				// Begin transfer

			}

		default:
			newMsg := []byte("Bro what? Unknown initial type.")
			newMsg = append([]byte{InitialTypeTextMessage}, newMsg...)
			_ = ConnWriteMessage(conn, newMsg)
			_ = conn.Close()
			return nil
		}

	}

	return nil
}

func ConnWriteMessage(conn *websocket.Conn, message []byte) error {
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		utils.LogData("E:Writing to websocket. Message was", string(message))
		return err
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
