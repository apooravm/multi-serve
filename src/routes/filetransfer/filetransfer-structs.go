package filetransfer

import (
	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
)

var (
	// File transfer map of all ongoing transfers
	FTMap *utils.ClientsMap[FTMeta] = utils.NewClientsMap[FTMeta]()

	// File transfer unique code generator
	FTCodeGenerator = FTCode_Generator{
		start_ID: 1,
	}
	// UserMap            *ClientsMap   = NewClientsMap()
	// SubscribedUsersMap *ClientsMap   = NewClientsMap()
	// ConnUpgrader = websocket.Upgrader{}

	ConnUpgrader                           = websocket.Upgrader{}
	FTUsersMap   *utils.ClientsMap[FTMeta] = utils.NewClientsMap[FTMeta]()
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

// Datasize uint16 -> max size would be 65kb
// Datasize uint32 -> max size would be 4.2gb
type FTPacket struct {
	Version  uint8
	Code     uint8
	DataSize uint16
	Data     []byte
}

// packetBytes := []byte{
//         1,                    // Version
//         2,                    // UniqueCode
//         0, 0, 0, 10,          // DataSize (4 bytes for uint32)
//         1,                    // IsSenderReceiver (1 byte)
//         'S', 'e', 'n', 'd', 'e', 'r', 0, // SenderReceiverName (null-terminated)
//         'H', 'e', 'l', 'l', 'o', // Data (5 bytes)
//     }
//
