package filetransfer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	pass = 1
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

// Since the 2 sockets are working concurrently
// Need to signal using a channel
// Work on it later
func FileTransferWs(c echo.Context) error {
	params := c.QueryParams()
	intentArgs := params["intent"]

	// Client intent can be either send or receive.
	var intent string
	var ongoingFTCache *FTMeta

	if len(intentArgs) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("No intent found. Must be send/receive."))
	}

	intent = intentArgs[0]
	var initialError error = nil

	// Handle initial metadata from both sender and receiver here
	switch intent {
	case "send":
		var fileinfo, sendername []string
		var allFileInfo []FileInfo

		fileinfo = params["fileinfo"]
		sendername = params["sendername"]

		if len(fileinfo) == 0 || len(sendername) != 1 {
			initialError = fmt.Errorf("Incomplete send data. Need fileinfo, sendername.")
			break
		}

		for _, info := range fileinfo {
			fparts := strings.Split(info, ",")
			if len(fparts) != 3 {
				initialError = fmt.Errorf("Invalid fileinfo syntax. Could not split fileinfo into path, size and id. %s", info)
				break
			}
			fpath, fsize, fId := fparts[0], fparts[1], fparts[2]

			parsedSize, err := strconv.ParseUint(fsize, 10, 64)
			if err != nil {
				initialError = fmt.Errorf("Invalid file size. Must be uint64. %s", fparts[0])
				break
			}

			parsedId, err := strconv.ParseUint(fId, 10, 8)
			if err != nil {
				initialError = fmt.Errorf("Invalid file Id. Must be uint8. %s", fparts[0])
				break
			}

			allFileInfo = append(allFileInfo, FileInfo{
				Name:         filepath.Base(fpath),
				RelativePath: fpath,
				Size:         parsedSize,
				Id:           uint8(parsedId),
			})
		}

		// Generate a uint8 unique code for the transfer.
		unique_code := FTCodeGenerator.NewCode()

		ongoingFTCache = &FTMeta{
			FileInfo:   &allFileInfo,
			SenderName: sendername[0],
			Code:       unique_code,
		}

	case "receive":
		var incomingRecvCode, receivername []string

		incomingRecvCode = params["code"]
		receivername = params["receivername"]

		if len(incomingRecvCode) != 1 || len(receivername) != 1 {
			// return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Incomplete receiver data. Need code, receivername."))
			initialError = fmt.Errorf("Incomplete receiver data. Need code, receivername.")
			break
		}

		parsedRecvCode, err := strconv.ParseUint(incomingRecvCode[0], 10, 8)
		if err != nil {
			initialError = fmt.Errorf("Invalid code. Must be uint8.")
			break
		}

		// Check if transfer exists
		ongoingFT, exists := FTMap.GetClient(fmt.Sprint(parsedRecvCode))
		if !exists {
			// return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Unknown code. Transfer not found."))
			initialError = fmt.Errorf("Unknown code. Transfer not found.")
			break
		}

		// Update the map with conn later
		ongoingFT.ReceiverName = receivername[0]
		ongoingFTCache = ongoingFT

	default:
		// return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid intent. Must be send/receive."))
		initialError = fmt.Errorf("Invalid intent. Must be send/receive.")
		break
	}

	// var isSender bool = false
	var err error
	ConnUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := ConnUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("WS_ERR:", err.Error())
		return err
	}

	// Incase error
	if initialError != nil {
		// reason for disconnect
		_ = ConnWriteMessage(conn, initialError.Error())

		// notify client about disconnect
		closeNotifPacket, err := CreateBinaryPacket(Version, InitialTypeCloseConnNotify)
		if err != nil {
			fmt.Printf("Could not create packet. %s\n", err.Error())

		} else {
			_ = conn.WriteMessage(websocket.BinaryMessage, closeNotifPacket)
		}

		_ = conn.Close()
		return nil
	}

	// If intent -> sender add to map else if intent -> receiver update the map
	if intent == "send" {
		ongoingFTCache.SenderConn = conn
		ongoingFTCache.stopCh = make(chan struct{})
		FTMap.AddClient(fmt.Sprint(ongoingFTCache.Code), ongoingFTCache)

		// Reply back to the sender with the code
		MDPacket, err := CreateBinaryPacket(Version, InitialTypeTransferCode, ongoingFTCache.Code)
		if err != nil {
			fmt.Println(err.Error())
		}

		if err := ongoingFTCache.SenderConn.WriteMessage(websocket.BinaryMessage, MDPacket); err != nil {
			fmt.Println("E:Writing to receiver. Transfer code response.", err.Error())
		}

	} else if intent == "receive" {
		ongoingFTCache.ReceiverConn = conn
		FTMap.UpdateClient(fmt.Sprint(ongoingFTCache.Code), func(client *FTMeta) *FTMeta {
			return ongoingFTCache
		})

		// Notify the sender about the receiver connecting
		senderPkt, err := CreateBinaryPacket(Version, InitialTypeTextMessage, []byte("Receiver has connected."))
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if err := ongoingFTCache.SenderConn.WriteMessage(websocket.BinaryMessage, senderPkt); err != nil {
				fmt.Println("E:Responding to sender with receiver joining.", err.Error())
			}
		}

		// Reply back to the receiver with the MD
		jsonByteArr, err := json.Marshal(*ongoingFTCache.FileInfo)
		if err != nil {
			if err := ConnWriteMessage(conn, "Could not marshal response metadata.", err.Error()); err != nil {
				fmt.Println("E:Writing to receiver. Metadata response marshalling error.", err.Error())
			}
			DisconnectBoth(intent, ongoingFTCache, "Internal server error. Could not encode metadata for receiver.")
			return nil
		}

		MDPacket, err := CreateBinaryPacket(Version, InitialTypeReceiverMD, jsonByteArr)
		if err != nil {
			fmt.Println(err.Error())
		}

		if err := ongoingFTCache.ReceiverConn.WriteMessage(websocket.BinaryMessage, MDPacket); err != nil {
			fmt.Println("E:Writing to receiver. Metadata response.", err.Error())
		}
	}

	// Only valid connections make it till here so no need for edge cases
	// ongoingFTCache should never be empty
	// defer DisconnectBoth(intent, ongoingFTCache)

	// Work on both single file transfers and transerring an entire dir
	// Sender
	// sender -> server setup FT [done]
	// server -> sender respond with code
	// receiver -> server respond with code and update FT. [done]
	// server -> receiver
	for {
		select {
		case <-ongoingFTCache.stopCh:
			fmt.Println("CHANNEL FINALLY WORKING???")
			return nil

		default:
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				ongoing, _ := FTMap.GetClient(fmt.Sprint(ongoingFTCache.Code))
				// Stopped by other conn
				if ongoing.TransferStopped {
					FTMap.DeleteClient(fmt.Sprint(ongoingFTCache.Code))
					return nil
				}

				var disconnReason string
				if intent == "send" {
					disconnReason = "Sender left."
				} else {
					disconnReason = "Receiver left."
				}

				_ = DisconnectBoth(intent, ongoingFTCache, disconnReason)
				return nil
			}

			if messageType != websocket.BinaryMessage {
				var disconnReason string
				if intent == "send" {
					disconnReason = "Invalid message type from sender."
				} else {
					disconnReason = "Invalid message type from receiver."
				}
				DisconnectBoth(intent, ongoingFTCache, disconnReason)
				return nil
			}

			// Only Version or less was sent
			if len(message) < 2 {
				// ConnWriteMessage(conn, []byte("Empty payload."))
				continue
			}

			switch message[1] {
			// from receiver
			case InitialTypeStartTransferWithId:
				if len(message) < 3 {
					ConnWriteMessage(conn, "No Id provided.")
					continue
				}
				updatedOngoingCache, _ := FTMap.GetClient(fmt.Sprint(ongoingFTCache.Code))
				ongoingFTCache = updatedOngoingCache

				if err := ongoingFTCache.SenderConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
					utils.LogData("Could not create start transfer with file Id packet for sender.", err.Error())
				}

			// from sender
			case InitialTypeTransferPacket:
				if len(message) < 3 {
					_ = ConnWriteMessage(conn, "Empty packet.")
					continue
				}

				if err := ongoingFTCache.ReceiverConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
					utils.LogData("Could not write to receiver with file packet.", err.Error())
				}

			// from receiver
			case InitialTypeRequestNextPacket:
				if err := ongoingFTCache.SenderConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
					utils.LogData("Could not request packet from sender.", err.Error())
				}

			case InitialTypeSingleFileTransferFinish:
				if err := ongoingFTCache.ReceiverConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
					utils.LogData("E:Could not write single transfer finish ping to receiver.")
				}

			case InitialTypeAllTransferFinish:
				if err := ongoingFTCache.ReceiverConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
					utils.LogData("E:Could not write all transfer finish ping to receiver.")
				}

				_ = DisconnectBoth(intent, ongoingFTCache, "Transfer finished.")
				return nil

			// Could be from sender, for now from receiver
			case InitialAbortTransfer:
				if intent == "send" {
					fmt.Println("TODO: Abort for sender.")

				} else {
					_ = DisconnectBoth(intent, ongoingFTCache, "Receiver aborted the transfer.")
					return nil

				}
			}

		}
	}

	// Memoizing the receiver conn.
	// For the first packet, sender conn looks up from the map
	// Saves to this and uses this for the remaining packets.
	// var ongoingFTCache *FTMeta

	// client_id := strconv.Itoa(utils.Id_Gen.GenerateNewID())
	// for {
	// 	// Sudden disconnect makes this throw err
	// 	// Incase either suddenly disconnect
	// 	// Send a message to the other and close the connection.
	// 	messageType, message, err := conn.ReadMessage()
	// 	if err != nil {
	// 		// Checking for exists should help in the ongoing loop of other conn
	// 		utils.LogData("E:Could not read from socket.", err.Error())
	//
	// 		ongoingFT, _ := FTMap.GetClient(string(ongoingFTCache.Code))
	// 		if ongoingFT.ReceiverClosed || ongoingFT.SenderClosed {
	// 			return nil
	// 		}
	//
	// 		// No other party (receiver) had connected. Can directly close sender.
	// 		if ongoingFT.ReceiverConn == nil {
	// 			_ = DisconnectConn(conn, "No response.")
	// 			FTMap.DeleteClient(string(ongoingFT.Code))
	// 			return nil
	// 		}
	//
	// 		DisconnectClient(conn, ongoingFTCache.Code, "Closing", "Closing")
	//
	// 		return nil
	// 	}
	//
	// 	if messageType != websocket.BinaryMessage {
	// 		utils.LogData("Unexpected message type.", strconv.Itoa(messageType))
	// 		ConnWriteMessage(conn, []byte("Unexpected message type"))
	// 		continue
	// 	}
	//
	// 	// only Version and initial_byte were sent
	// 	if len(message) < 1 {
	// 		fmt.Println(message)
	// 		ConnWriteMessage(conn, []byte("Empty payload."))
	// 		continue
	// 	}
	//
	// 	// first byte always should be the Version
	// 	// second the initial_byte
	// 	// [Version][initial_byte][...]
	// 	// Different initial types
	// 	switch message[1] {
	// 	case InitialTypeRegisterSender:
	// 		// isSender = true
	// 		newFT, err := HandleRegisterSender(message, conn)
	// 		if err != nil {
	// 			DisconnectConn(conn, err.Error())
	// 			return nil
	// 		}
	// 		ongoingFTCache = newFT
	//
	// 	case InitialTypeRegisterReceiver:
	// 		incomingBuffer := bytes.NewBuffer(message[2:])
	//
	// 		code, recvName, err := ParseReceiverCode(incomingBuffer)
	// 		if err != nil {
	// 			_ = DisconnectConn(conn, "Could not parse code.")
	// 			return nil
	// 		}
	//
	// 		_, exists := FTMap.GetClient(string(code))
	// 		if !exists {
	// 			if err := DisconnectConn(conn, "Sender not found."); err != nil {
	// 				fmt.Println(err.Error())
	// 			}
	// 			return nil
	// 		}
	//
	// 		newFT, resp, err := HandleRegisterReceiver(recvName, code, incomingBuffer, conn)
	// 		if err != nil {
	// 			DisconnectConn(conn, resp.ReceiverDiscnMsg)
	// 			DisconnectConn(newFT.SenderConn, resp.ReceiverDiscnMsg)
	// 			return nil
	// 		}
	// 		ongoingFTCache = newFT
	//
	// 	// Sender -> Recv
	// 	case InitialTypeTransferPacket:
	// 		ongoig, _ := FTMap.GetClient(string(ongoingFTCache.Code))
	// 		ongoingFTCache = &ongoig
	//
	// 		pass += 1
	// 		if err := ongoingFTCache.ReceiverConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
	// 			fmt.Println("E:Requesting next packet")
	// 		}
	//
	// 	// Recv -> Sender
	// 	case InitialTypeRequestNextPkt:
	// 		if err := ongoingFTCache.SenderConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
	// 			fmt.Println("E:Requesting next packet")
	// 		}
	//
	// 	// Sender -> Server
	// 	case InitialTypeFinishTransfer:
	// 		finishPing := append([]byte{Version, InitialTypeFinishTransfer}, message...)
	// 		if err := ongoingFTCache.ReceiverConn.WriteMessage(websocket.TextMessage, finishPing); err != nil {
	// 			utils.LogData("E:Writing to websocket. Message was", string(message))
	// 		}
	//
	// 		DisconnectClient(conn, ongoingFTCache.Code, "", "Transfer finished")
	//
	// 		// Packet frame
	// 	// 0 -> sender, 1 -> receiver
	// 	// [init_byte][unique_code][sender_or_reic]
	// 	// May not need this one here
	// 	case InitialTypeCloseConn:
	// 		return nil
	//
	// 	case InitialAbortTransfer:
	// 		if ongoingFTCache.SenderConn == conn {
	// 			DisconnectClient(conn, ongoingFTCache.Code, "", "Sender aborted")
	//
	// 		} else {
	// 			DisconnectClient(conn, ongoingFTCache.Code, "Receiver aborted", "")
	// 		}
	//
	// 		return nil
	//
	// 	default:
	// 		DisconnectConn(conn, "Bro what? Unknown type.")
	// 		return nil
	// 	}
	//
	// }
}

func HandleSender(ongoingFT *FTMeta) error {

	return nil
}

func HandleReceiver(intent string, ongoingFT *FTMeta) error {

	return nil
}

// func WriteBinaryPkt(conn *websocket.Conn, packet []byte, errorMessage string, isSender bool) error {
// 	if err := conn.WriteMessage(websocket.BinaryMessage, packet); err != nil {
// 		DisconnectClient(errorMessage, conn, isSender)
// 		return err
// 	}
//
// 	return nil
// }

// Handling disconnection at the top level
// Funcs just send a cust error
func HandleRegisterSender(message []byte, conn *websocket.Conn) (*FTMeta, error) {
	var clientHandshake ClientHandshake

	// Ignore the Version and initial byte
	if err := json.Unmarshal(message[2:], &clientHandshake); err != nil {
		utils.LogData("E:Unmarshalling json response.", err.Error())
		// DisconnectClient("Internal server error. Decoding Json.", conn, isSender)
		return nil, err
	}

	// This could be done better.
	// Only 255 unique ones possible
	unique_code := FTCodeGenerator.NewCode()

	resp, err := CreateBinaryPacket(Version, InitialTypeUniqueCode, unique_code)
	if err != nil {
		utils.LogData("E:Creating binary packet.", err.Error())
		// DisconnectClient("Internal server error. Creating binary packet.", conn, isSender)
		return nil, err
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, resp); err != nil {
		utils.LogData("E:Could not send unique code")
		// DisconnectClient("Could not write packet to connection.", conn, isSender)
		return nil, err
	}

	// Says client but actually is the whole process
	var newFTMeta FTMeta
	newFTMeta.Version = Version
	newFTMeta.Code = unique_code
	newFTMeta.SenderConn = conn
	newFTMeta.SenderName = clientHandshake.ClientName
	newFTMeta.FileSize = clientHandshake.FileSize
	newFTMeta.Filename = clientHandshake.Filename

	FTMap.AddClient(string(unique_code), &newFTMeta)

	return &newFTMeta, nil
}

func ParseReceiverCode(incomingBuffer *bytes.Buffer) (uint8, string, error) {
	// Receiver packet
	// [Version][initial_byte][unique_code][receiver_name]
	var incomingReceiverCode uint8

	if err := binary.Read(incomingBuffer, binary.BigEndian, &incomingReceiverCode); err != nil {
		utils.LogData("E:Reading binary packet, code.", err.Error())
		return 0, "", err
	}

	incomingReceiverNameBytes, err := io.ReadAll(incomingBuffer)
	if err != nil {
		utils.LogData("E:Reading binary packet, name.", err.Error())
		// DisconnectClient("Could not read name from receiver.", conn)
		return 0, "", err
		// return fmt.Errorf("Could not read name from receiver.")
	}

	incomingReceiverName := string(incomingReceiverNameBytes)

	return incomingReceiverCode, incomingReceiverName, nil
}

func HandleRegisterReceiver(incomingRecvName string, incomingCode uint8, incomingBuffer *bytes.Buffer, conn *websocket.Conn) (*FTMeta, *FTErrResp, error) {
	// the read position in the buffer changes with every read
	// this just reads everything from the current position till the end
	// Update the FTMeta info with the receiver_name and conn

	FTMap.UpdateClient(string(incomingCode), func(client *FTMeta) *FTMeta {
		client.ReceiverConn = conn
		client.ReceiverName = incomingRecvName
		return client
	})

	// Update the newFTMeta for the current client
	ongoingFT, _ := FTMap.GetClient(string(incomingCode))

	resp, err := CreateBinaryPacket(Version, InitialTypeTextMessage)
	if err != nil {
		utils.LogData("E:Creating binary packet. Notifying sender about receiver connecting.", err.Error())
		return nil, &FTErrResp{
			ReceiverDiscnMsg: "Internat server error. Could not create binary packet.",
			SenderDiscnMsg:   "Internat server error. Could not create binary packet.",
		}, err
	}

	resp = append(resp, []byte("Receiver has connected.")...)
	if err := ongoingFT.SenderConn.WriteMessage(websocket.BinaryMessage, resp); err != nil {
		// Disconnect not necessary
	}

	// After the FTMap has been updated with the receiver's data
	// TransferMD packet frame
	// [Version][initial_byte][json_transferMD]

	// Responding with the transfers Metadata
	transferMD := MDReceiver{
		FileSize:   ongoingFT.FileSize,
		SenderName: ongoingFT.SenderName,
		Filename:   ongoingFT.Filename,
	}

	resByteArr, err := json.Marshal(&transferMD)
	if err != nil {
		utils.LogData("E:filetransfer Could not marshal metadata response.", err.Error())
		return nil, &FTErrResp{
			ReceiverDiscnMsg: "Internat server error. Could not encode json data.",
			SenderDiscnMsg:   "Internat server error. Could not encode json data.",
		}, err
	}

	resByteArr = append([]byte{Version, InitialTypeTransferMetaData}, resByteArr...)
	if err := conn.WriteMessage(websocket.BinaryMessage, resByteArr); err != nil {
		utils.LogData("E:filetransfer Could not write to receiver.", err.Error())
		return nil, &FTErrResp{
			ReceiverDiscnMsg: "Could not write.",
			SenderDiscnMsg:   "Could not write to receiver.",
		}, err
	}

	return ongoingFT, nil, nil
}

// Handle Client socket disconnection
// Graceful handling prevents error logs
func DisconnectClient(currConn *websocket.Conn, code uint8, senderMessage, recvMessage string) {
	// Check if client is a part of any transfer in newFTMeta
	// Disconnect other client if so
	ongoingFT, exists := FTMap.GetClient(string(code))
	// Single disconnect?
	if !exists {
		DisconnectConn(currConn, "")
		return
	}

	if ongoingFT.ReceiverClosed && ongoingFT.SenderClosed {
		return
	}

	err := DisconnectConn(ongoingFT.SenderConn, senderMessage)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = DisconnectConn(ongoingFT.ReceiverConn, recvMessage)
	if err != nil {
		fmt.Println(err.Error())
	}

	FTMap.UpdateClient(string(code), func(client *FTMeta) *FTMeta {
		client.ReceiverClosed = true
		client.SenderClosed = true
		return client
	})

	FTMap.DeleteClient(string(code))
}

// Server sends closeConn byte
// Issue is, both the connections have an infinite loop going on
// Closing one conn doesnt stop the loop
// Unless I make changes to the structs themselves there will always will be errors
func DisconnectConn(conn *websocket.Conn, message string) error {
	if conn == nil {
		return nil
	}

	var newMsg []byte

	// First send a text message type with the message, if any
	if message != "" {
		newMsg = []byte(message)
		newMsg = append([]byte{Version, InitialTypeTextMessage}, newMsg...)
		_ = conn.WriteMessage(websocket.BinaryMessage, newMsg)
	}

	// Second, send another empty frame with just the InitialTypeCloseConn init_byte
	newMsg = []byte{Version, InitialTypeCloseConn}

	if err := conn.WriteMessage(websocket.TextMessage, newMsg); err != nil {
		utils.LogData("E:Writing to websocket. Message was", string(message))
		return fmt.Errorf("E:Writing to socket. %s %s", message, err.Error())
	}

	// Little delay since sometimes the closeConn doesnt get written to the conn
	time.Sleep(50 * time.Millisecond)
	err := conn.Close()
	if err != nil {
		utils.LogData(err.Error())
		return fmt.Errorf("E:Writing closing socket. %s %s", message, err.Error())
	}

	return nil
}
