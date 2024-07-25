package filetransfer

import (
	"github.com/gorilla/websocket"
)

var (
	// UserMap            *ClientsMap   = NewClientsMap()
	// SubscribedUsersMap *ClientsMap   = NewClientsMap()
	// ConnUpgrader = websocket.Upgrader{}

	ConnUpgrader = websocket.Upgrader{}
)

// packetBytes := []byte{
//         1,                    // Version
//         2,                    // UniqueCode
//         0, 0, 0, 10,          // DataSize (4 bytes for uint32)
//         1,                    // IsSenderReceiver (1 byte)
//         'S', 'e', 'n', 'd', 'e', 'r', 0, // SenderReceiverName (null-terminated)
//         'H', 'e', 'l', 'l', 'o', // Data (5 bytes)
//     }
//
