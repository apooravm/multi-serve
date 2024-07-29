package filetransfer

import (
	"fmt"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
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

	InitialTypeRequestNextPkt = uint8(0x10)
	InitialTypeFinishTransfer = uint8(0x11)

	InitialAbortTransfer = uint8(0x12)

	// current version
	version = byte(1)
)

var (
	// UserMap            *ClientsMap   = NewClientsMap()
	// SubscribedUsersMap *ClientsMap   = NewClientsMap()
	// ConnUpgrader = websocket.Upgrader{}
	// File transfer map of all ongoing transfers
	FTMap *utils.ClientsMap[FTMeta] = utils.NewClientsMap[FTMeta]()
	// File transfer unique code generator
	FTCodeGenerator = FTCode_Generator{
		start_ID: 1,
	}

	FTUsersMap *utils.ClientsMap[FTMeta] = utils.NewClientsMap[FTMeta]()

	ConnUpgrader = websocket.Upgrader{}
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
	Code           uint8
	SenderName     string
	ReceiverName   string
	SenderConn     *websocket.Conn
	ReceiverConn   *websocket.Conn
	Filename       string
	FileSize       uint64
	FileInfo       *[]FileInfo
	Version        uint8
	SenderClosed   bool
	ReceiverClosed bool
	stopCh         chan struct{}
}

type FileInfo struct {
	Name string
	// Relative to the target folder.
	RelativePath string
	// Abs path of the file in the system.
	AbsPath string
	Size    uint64
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

// packetBytes := []byte{
//         1,                    // Version
//         2,                    // UniqueCode
//         0, 0, 0, 10,          // DataSize (4 bytes for uint32)
//         1,                    // IsSenderReceiver (1 byte)
//         'S', 'e', 'n', 'd', 'e', 'r', 0, // SenderReceiverName (null-terminated)
//         'H', 'e', 'l', 'l', 'o', // Data (5 bytes)
//     }
//

// Object returned with error containing messages for sender, receiver
type FTErrResp struct {
	Simple           string
	SenderDiscnMsg   string
	ReceiverDiscnMsg string
}
