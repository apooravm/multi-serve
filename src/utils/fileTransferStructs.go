package utils

import "github.com/gorilla/websocket"

// FileTransfer
// -------
/*
Receiver pings the server with the IP and port
Server generates a unique connection ID String len = 4
Receiver getsthe ID and the user may share the ID

type FTPayload struct {
	name string => receiver | sender
	type string => ip, id, metadata
	payload interface{}
}

Rec side
type FTPayload struct {
	type string => id, metadata
	payload interface{}
}

send side
type FTPayload struct {
	type string => [ip, port]
	payload interface{}
}

// consts for the type of payload
// Metadata from the sender will include the ID
const (
	METADATA = "metadata"
	IP_PORT = "ip_port"
	ID = "id"
)

Rec - [IP, Port]
Serv-Rec - ID
Rec-Send - ID (private)
Send - [ID, Metadata]
Serv-Send - [IP, Port]
Serv-Rec - Metdata

Sender pings the server with the conn ID and the MetaData and receives back the IP and port

Maybe use websockets instead???
w ws

Rec => server Message
type Message struct {
	Sender    string
	Direction string
	Config    string
	Content   string
	Password  string
}
if Sender == "receiver"
*/

// consts for the type of payload
// Metadata from the sender will include the ID

const (
	METADATA   = "metadata"
	RECEIV_URL = "receiverURL"
)

// type FTPayload struct {
// 	name string => receiver | sender
// 	type string => ip, id, metadata
// 	payload interface{}
// }

type FTPayload struct {
	SentBy  string
	Command string
	Content interface{}
}

// All of the metadata
// recerverURL, connID, filename, fileByteSize
type FT_MetaData_plusCONN struct {
	ReceiverUrl  string
	ConnID       string
	Filename     string
	FileByteSize int64

	ReceiverConn *websocket.Conn
	SenderConn   *websocket.Conn
}

type MetaDataFmt struct {
	ConnID       string
	Filename     string
	FileByteSize int64
}

type FT_Transfer struct {
	ReceiverConn *websocket.Conn
	SenderConn   *websocket.Conn

	ReceiverApproved bool
	TransferID       string

	Filename     string
	FileByteSize int64
}

var (
	FT_Map = make(map[string]*FT_Transfer)
)
