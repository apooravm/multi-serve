package filetransfer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
)

// Creates a binry packet with the given components. Byte order big endian
func CreateBinaryPacket(parts ...any) ([]byte, error) {
	responseBfr := new(bytes.Buffer)
	for _, part := range parts {
		if err := binary.Write(responseBfr, binary.BigEndian, part); err != nil {
			return nil, err
		}
	}

	return responseBfr.Bytes(), nil
}

// Writes a string message to the conn with the InitialTypeTextMessage byte.
func ConnWriteMessage(conn *websocket.Conn, message ...string) error {
	joinedMsg := strings.Join(message, " ")
	byteMsg := []byte(joinedMsg)

	finalMsg := append([]byte{version, InitialTypeTextMessage}, byteMsg...)
	if err := conn.WriteMessage(websocket.TextMessage, finalMsg); err != nil {
		utils.LogData("E:Writing to websocket. Message was", joinedMsg)
		return err
	}

	return nil
}

// Handles the disconnection for both sender and receiver.
// No need for individual disconnection for now.
// If sender -> If recv not connected: delete map and return err
// If recv connected: disconnnect recv through channel, delete map and return
// If Receiver -> Sender always connected: disconnect sender through channel, delete map and return
// Channel not working for some reason.
// Instead just updating the TransferStopped flag in the FTMeta and
// checking if it was when the other conn crashes.
// Other conn then deletes the obj from map and returns
func DisconnectBoth(intent string, ongoingFT *FTMeta) error {
	updatedOngoingFT, _ := FTMap.GetClient(fmt.Sprint(ongoingFT.Code))

	FTMap.UpdateClient(fmt.Sprint(updatedOngoingFT.Code), func(client FTMeta) FTMeta {
		client.TransferStopped = true
		return client
	})

	// Sent to notify the client of the imminent disconnect.
	closeNotifPacket, _ := CreateBinaryPacket(version, InitialTypeCloseConnNotify)

	if intent == "send" {
		// If recv connected
		// Notify the connected ones about the closing.
		if err := updatedOngoingFT.SenderConn.WriteMessage(websocket.BinaryMessage, closeNotifPacket); err != nil {
			utils.LogData("E:Writing notif to sender.", err.Error())
		}

		if err := updatedOngoingFT.SenderConn.Close(); err != nil {
			utils.LogData("E:Closing sender conn.", err.Error())
		}

		if updatedOngoingFT.ReceiverConn != nil {
			disconnReason := "Sender left."
			// Ignoring err for packet creation since its always going to work
			pkt, _ := CreateBinaryPacket(version, InitialTypeTextMessage, []byte(disconnReason))
			_ = updatedOngoingFT.ReceiverConn.WriteMessage(websocket.BinaryMessage, pkt)

			_ = updatedOngoingFT.ReceiverConn.WriteMessage(websocket.BinaryMessage, closeNotifPacket)

			if err := updatedOngoingFT.ReceiverConn.Close(); err != nil {
				utils.LogData("E:Closing receiver conn.", err.Error())
			}

			// If receiver has not connected yet, delete here directly.
		} else {
			FTMap.DeleteClient(fmt.Sprint(updatedOngoingFT.Code))

		}

	} else if intent == "receive" {
		disconnReason := "Receiver left."
		pkt, _ := CreateBinaryPacket(version, InitialTypeTextMessage, []byte(disconnReason))
		_ = updatedOngoingFT.SenderConn.WriteMessage(websocket.BinaryMessage, pkt)

		_ = updatedOngoingFT.SenderConn.WriteMessage(websocket.BinaryMessage, closeNotifPacket)
		if err := updatedOngoingFT.SenderConn.Close(); err != nil {
			utils.LogData("E:Closing sender conn.", err.Error())
		}

		_ = updatedOngoingFT.ReceiverConn.WriteMessage(websocket.BinaryMessage, closeNotifPacket)
		if err := updatedOngoingFT.ReceiverConn.Close(); err != nil {
			utils.LogData("E:Closing receiver conn.", err.Error())
		}
	}

	close(updatedOngoingFT.stopCh)
	return nil
}
