package filetransfer

import (
	"bytes"
	"encoding/binary"
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
func DisconnectBoth(intent string, ongoingFT *FTMeta, conn *websocket.Conn) error {
	if ongoingFT.SenderConn == nil && ongoingFT.ReceiverConn == nil {
		conn.Close()
		return nil
	}
	// All this in its own func
	// If sender -> If recv not connected: delete map and return err
	// If recv connected: disconnnect recv through channel, delete map and return
	// If Receiver -> Sender always connected: disconnect sender through channel, delete map and return

	if intent == "send" {
		// If recv connected
		// Notify the connected ones about the closing.
		closeNotifPacket, _ := CreateBinaryPacket(version, InitialTypeCloseConnNotify)
		_ = ongoingFT.SenderConn.WriteMessage(websocket.BinaryMessage, closeNotifPacket)

		if ongoingFT.ReceiverConn != nil {
			_ = ongoingFT.ReceiverConn.WriteMessage(websocket.BinaryMessage, closeNotifPacket)
			close(ongoingFT.stopCh)
		}

	} else if intent == "receive" {
		closeNotifPacket, _ := CreateBinaryPacket(version, InitialTypeCloseConnNotify)
		_ = ongoingFT.SenderConn.WriteMessage(websocket.BinaryMessage, closeNotifPacket)
		_ = ongoingFT.ReceiverConn.WriteMessage(websocket.BinaryMessage, closeNotifPacket)
		close(ongoingFT.stopCh)
	}

	FTMap.DeleteClient(string(ongoingFT.Code))
	return nil
}
