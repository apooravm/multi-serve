package utils

import (
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	// File transfer map of all ongoing transfers
	FTMap *ClientsMap[FTMeta] = NewClientsMap[FTMeta]()
	// File transfer unique code generator
	FTCodeGenerator = FTCode_Generator{
		start_ID: 1,
	}

	FTUsersMap *ClientsMap[FTMeta] = NewClientsMap[FTMeta]()
)

type FTCode_Generator struct {
	start_ID uint8
}

// Just increments the default value
func (idGen *FTCode_Generator) NewCode() uint8 {
	ret_id := idGen.start_ID
	idGen.start_ID += 1

	return ret_id
}

// Metadata for single file transfer transaction.
type FTMeta struct {
	Code         uint8
	SenderName   string
	ReceiverName string
	SenderConn   *websocket.Conn
	ReceiverConn *websocket.Conn
	Filename     string
	FileSize     uint64
	Version      uint8
}

func (ft *FTMeta) DisconnectBoth(messageSender string, messageReceiver string) {
	err := ft.SenderConn.Close()
	if err != nil {
		fmt.Println("Err closing sender conn", err.Error())
	}
	err = ft.ReceiverConn.Close()
	if err != nil {
		fmt.Println("Err closing receiver conn", err.Error())
	}
}

// Datasize uint16 -> max size would be 65kb
// Datasize uint32 -> max size would be 4.2gb
type FTPacket struct {
	Version  uint8
	Code     uint8
	DataSize uint16
	Data     []byte
}
