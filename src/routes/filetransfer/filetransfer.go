package filetransfer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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
	newFTMeta utils.FTMeta
	// Makes disconnections easier
	suddenDisconnection = false
	connClosed          = false
	unique_code         uint8
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
	var isSender bool = false
	var err error
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
			// Checking for exists should help in the ongoing loop of other conn
			fmt.Println("Someone disconnected")
			fmt.Println(isSender)
			ongoingFT, exists := utils.FTMap.GetClient(string(unique_code))
			if !exists {
				return nil
			}
			utils.LogData("E:Could not read from socket.", err.Error())
			if isSender {
				fmt.Println("Sender left")
				// DisconnectConn(conn, "")
				if ongoingFT.ReceiverConn != nil {
					DisconnectConn(ongoingFT.ReceiverConn, "Sender left.")
				}

				utils.FTMap.DeleteClient(string(ongoingFT.Code))

				return nil
			}

			fmt.Println("Receiver left")

			// DisconnectConn(conn, "")
			if ongoingFT.SenderConn != nil {
				DisconnectConn(ongoingFT.SenderConn, "Receiver left.")
			}

			utils.FTMap.DeleteClient(string(ongoingFT.Code))

			break
		}

		if messageType != websocket.BinaryMessage {
			utils.LogData("Unexpected message type.", strconv.Itoa(messageType))
			DisconnectClient("Unexpected message type.", conn, isSender)
			break
		}

		// only version and initial_byte were sent
		if len(message) < 2 {
			DisconnectClient("Empty Payload.", conn, isSender)
			return nil
		}

		// first byte always should be the version
		// second the initial_byte
		// [version][initial_byte][...]
		// Different initial types
		switch message[1] {
		case InitialTypeRegisterSender:
			isSender = true
			if err := HandleRegisterSender(message, conn); err != nil {
				return nil
			}

		case InitialTypeRegisterReceiver:
			if err := HandleRegisterReceiver(message, conn); err != nil {
				return nil
			}

		case InitialTypeBeginTransfer:
			if len(message) != 3 {
				DisconnectClient("Invalid frame. Cannot begin transfer.", conn, isSender)
				return nil
			}

			var beginTransferOrNo uint8 = message[2]

			// Receiver somehow moves to this step without registering
			// Should not happen, but just in case.
			if newFTMeta.ReceiverConn == nil {
				DisconnectClient("Receiver not registered.", conn, isSender)
				return nil
			}

			resp, err := CreateBinaryPacket(version, InitialTypeBeginTransfer, beginTransferOrNo)
			if err != nil {
				utils.LogData("E:Creating binary packet to begin transfer.")
				DisconnectClient("Internal server error. Could not create packet.", conn, isSender)
			}

			// Really gotta do smn about these funcs
			// Getting annoying writing them again and again
			if err := newFTMeta.SenderConn.WriteMessage(websocket.BinaryMessage, resp); err != nil {
				utils.LogData("E:Writing message to sender.", err.Error())
				DisconnectClient("Could not write to sender.", conn, isSender)
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
			if isSender {
				DisconnectClient("Sender left.", conn, isSender)
			} else {
				DisconnectClient("Receiver left.", conn, isSender)
			}

			return nil

		default:
			DisconnectClient("Bro what? Unknown initial type.", conn, isSender)
			return nil
		}

	}

	return nil
}

// Make sure to append [vesion][initial_byte]
func ConnWriteMessage(conn *websocket.Conn, message []byte) error {
	finalMsg := append([]byte{version, InitialTypeTextMessage}, message...)
	if err := conn.WriteMessage(websocket.TextMessage, finalMsg); err != nil {
		utils.LogData("E:Writing to websocket. Message was", string(message))
		return err
	}

	return nil
}

func CreateBinaryPacket(parts ...any) ([]byte, error) {
	responseBfr := new(bytes.Buffer)
	for _, part := range parts {
		if err := binary.Write(responseBfr, binary.BigEndian, part); err != nil {
			return nil, err
		}
	}

	return responseBfr.Bytes(), nil
}

func WriteBinaryPkt(conn *websocket.Conn, packet []byte, errorMessage string, isSender bool) error {
	if err := conn.WriteMessage(websocket.BinaryMessage, packet); err != nil {
		DisconnectClient(errorMessage, conn, isSender)
		return err
	}

	return nil
}

func HandleRegisterSender(message []byte, conn *websocket.Conn, isSender bool) error {
	var clientHandshake ClientHandshake
	// Ignore the version and initial byte
	if err := json.Unmarshal(message[2:], &clientHandshake); err != nil {
		utils.LogData("E:Unmarshalling json response.", err.Error())
		DisconnectClient("Internal server error. Decoding Json.", conn, isSender)
		return err
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
	unique_code = utils.FTCodeGenerator.NewCode()

	resp, err := CreateBinaryPacket(version, InitialTypeUniqueCode, unique_code)
	if err != nil {
		utils.LogData("E:Creating binary packet.", err.Error())
		DisconnectClient("Internal server error. Creating binary packet.", conn, isSender)
		return err
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, resp); err != nil {
		utils.LogData("E:Could not send unique code")
		DisconnectClient("Could not write packet to connection.", conn, isSender)
		return err
	}

	// Says client but actually is the whole process
	newFTMeta.Version = version
	newFTMeta.Code = unique_code
	newFTMeta.SenderConn = conn
	newFTMeta.SenderName = clientHandshake.ClientName
	newFTMeta.FileSize = clientHandshake.FileSize
	newFTMeta.Filename = clientHandshake.Filename

	utils.FTMap.AddClient(string(unique_code), &newFTMeta)

	return nil
}

func HandleRegisterReceiver(message []byte, conn *websocket.Conn) error {
	// Receiver packet
	// [version][initial_byte][unique_code][receiver_name]
	incomingBuffer := bytes.NewReader(message[2:])
	var incomingReceiverCode uint8

	if err := binary.Read(incomingBuffer, binary.BigEndian, &incomingReceiverCode); err != nil {
		utils.LogData("E:Reading binary packet, code.", err.Error())
		return fmt.Errorf("Could not read code from receiver.")
	}

	// the read position in the buffer changes with every read
	// this just reads everything from the current position till the end
	incomingReceiverNameBytes, err := io.ReadAll(incomingBuffer)
	if err != nil {
		utils.LogData("E:Reading binary packet, name.", err.Error())
		// DisconnectClient("Could not read name from receiver.", conn)
		return fmt.Errorf("Could not read name from receiver.")
	}
	incomingReceiverName := string(incomingReceiverNameBytes)

	// Update the FTMeta info with the receiver_name and conn
	ongoingFT, exists := utils.FTMap.GetClient(string(incomingReceiverCode))
	if !exists {
		// DisconnectConn(conn, "Sender not found")
		return fmt.Errorf("Sender not found.")
	}

	utils.FTMap.UpdateClient(string(incomingReceiverCode), func(client utils.FTMeta) utils.FTMeta {
		client.ReceiverConn = conn
		client.ReceiverName = incomingReceiverName
		return client
	})

	// Update the newFTMeta for the current client
	ongoingFT, exists = utils.FTMap.GetClient(string(incomingReceiverCode))
	if exists {
		newFTMeta = ongoingFT
	}

	resp, err := CreateBinaryPacket(version, InitialTypeTextMessage)
	if err != nil {
		utils.LogData("E:Creating binary packet. Notifying sender about receiver connecting.", err.Error())
		return fmt.Errorf("Could not create binary packet.")
	}

	resp = append(resp, []byte("Receiver has connected.")...)
	if err := ongoingFT.SenderConn.WriteMessage(websocket.BinaryMessage, resp); err != nil {

	}

	// After the FTMap has been updated with the receiver's data
	// TransferMD packet frame
	// [version][initial_byte][json_transferMD]
	// transferMD := new(bytes.Buffer)
	//
	// if err := binary.Write(transferMD, binary.BigEndian, InitialTypeTransferMetaData); err != nil {
	// 	fmt.Println("Could not create binary response", err.Error())
	// 	_ = conn.Close()
	// 	return nil
	// }

	transferMD := MDReceiver{
		FileSize:   ongoingFT.FileSize,
		SenderName: ongoingFT.SenderName,
		Filename:   ongoingFT.Filename,
	}

	resByteArr, err := json.Marshal(&transferMD)
	if err != nil {
		fmt.Println("Could not marshal metadata response", err.Error())
		utils.LogData("E:filetransfer Could not marshal metadata response.", err.Error())
		DisconnectClient("Internal server error. Json encoding receiver data..", conn)
		return err
	}

	resByteArr = append([]byte{version, InitialTypeTransferMetaData}, resByteArr...)
	if err := conn.WriteMessage(websocket.BinaryMessage, resByteArr); err != nil {
		utils.LogData("E:filetransfer Could not write to receiver.", err.Error())
		DisconnectClient("Could not write to receiver.", conn)
		return err
	}

	return nil
}

// Handle Client socket disconnection
// Graceful handling prevents error logs
func DisconnectClient(disconnection_message string, conn *websocket.Conn, isSender bool) {
	// Check if client is a part of any transfer in newFTMeta
	// Disconnect other client if so

	if isSender {
		// Basically, sudden disconnection is a special one that happens only at 1 place
		// Instead of changing the disconnect func everywhere, I just check for the suddenDisconnection flag

		newMsg := []byte(disconnection_message)
		newMsg = append([]byte{version, InitialTypeCloseConn}, newMsg...)
		_ = ConnWriteMessage(conn, newMsg)
		err := conn.Close()
		if err != nil {
			fmt.Println(err.Error())
		}

		// If receiver connected, disconnct them
		if newFTMeta.ReceiverConn != nil {
			newMsg := []byte(disconnection_message)
			newMsg = append([]byte{version, InitialTypeCloseConn}, newMsg...)
			_ = ConnWriteMessage(newFTMeta.ReceiverConn, newMsg)
			err = newFTMeta.ReceiverConn.Close()
			if err != nil {
				fmt.Println("Receiver err", err.Error())
			}

		}

		utils.FTMap.DeleteClient(string(newFTMeta.Code))
		connClosed = true
		return
	}

	// If receiver
	newMsg := []byte(disconnection_message)
	newMsg = append([]byte{version, InitialTypeCloseConn}, newMsg...)
	_ = ConnWriteMessage(conn, newMsg)
	_ = conn.Close()

	// If sender connected, disconnct them
	// Sender should always be connected, just in case
	if newFTMeta.SenderConn != nil {
		newMsg := []byte(disconnection_message)
		newMsg = append([]byte{version, InitialTypeCloseConn}, newMsg...)
		_ = ConnWriteMessage(newFTMeta.SenderConn, newMsg)
		_ = newFTMeta.SenderConn.Close()

	}

	utils.FTMap.DeleteClient(string(unique_code))
	connClosed = true
}

// Server sends closeConn byte
// Issue is, both the connections have an infinite loop going on
// Closing one conn doesnt stop the loop
// Unless I make changes to the structs themselves there will always will be errors
func DisconnectConn(conn *websocket.Conn, message string) {
	var newMsg []byte
	if message != "" {
		newMsg = []byte(message)
		newMsg = append([]byte{version, InitialTypeCloseConn}, newMsg...)
	} else {
		newMsg = []byte{version, InitialTypeCloseConn}
	}

	err := ConnWriteMessage(conn, newMsg)
	if err != nil {
		fmt.Println(err.Error())
		utils.LogData(err.Error())
	}

	// Little delay since sometimes the closeConn doesnt get written to the conn
	// time.Sleep(0 * time.Millisecond)
	err = conn.Close()
	if err != nil {
		fmt.Println(err.Error())
		utils.LogData(err.Error())
	}
}
