package routes

import (
	"log"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

/*
- Rec asks for an ID from Server through a COMMAND object
- Server generates a new ID, creates a new TRANSFER object
  adds the ID and the rec conn to it. Then sends back the ID
- Rec user communicates the ID to the sender
- Sender creates a METADATA object and adds the filebytesize, filename, ID to it
  It then sends the objec to the server through a COMMAND object
- Server checks the command
- Checks whether the Map[ID] != nil
- If not, it sends the MetaData to the receiver
- The receiver confirms that it is ready for the data to the server
- The server then notifies the sender to send the packets
- The packets are sent through a PACKET object
- PACKET obj contains, the packet itself, size of the packet and the ID
- The server maps the packet to the receiver through the ID
- Ones the process is finished, the Sender sends a <END> flag in a PACKET obj
- Server relays the message to the rec
- Receiver stops receiving and writes the packets to file
-
*/

type FT_Command struct {
	SentBy     string
	TransferID string
	Command    string
	Payload    interface{}
}

type FT_MetaData struct {
	TransferID   string
	Filename     string
	FileByteSize int64
}

type FT_Packet struct {
	Payload    []byte
	Size       int
	TransferID string
}

func SendCommandToConn(command *FT_Command, conn *websocket.Conn) {
	if err := conn.WriteJSON(command); err != nil {
		log.Println("ERR FileTransfer: Could Not Write JSON to Conn")
	}
}

func gen_TransferID() string {
	return "0001"
}

func verify_TransferID(transferID string) bool {
	return utils.FT_Map[transferID] != nil
}

// FT_Command.Payload can be either

var (
	ErrInvalidFormat     = FT_Command{SentBy: "server", Command: "error", Payload: "Invalid Command Format"}
	ErrInvalidTransferID = FT_Command{SentBy: "server", Command: "error", Payload: "Invalid Transfer ID"}
)

func FileTransferGroup_old(group *echo.Group) {
	group.GET("", FileTransfer_old)
}

func FileTransfer_old(c echo.Context) error {
	conn, err := utils.ConnUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return &utils.ServerError{
			Err:    err,
			Code:   500,
			Simple: "FileTransfer: Failed to Upgrade HTTP To WebSocket",
		}
	}

	for {
		var command FT_Command
		if err := conn.ReadJSON(&command); err != nil {
			conn.Close()
			return nil
		}

		if command.SentBy == "receiver" {
			switch command.Command {
			case "gen_TransferID":
				// Generate ID
				// Create New Transfer Object in utils.FT_Map
				// Send the TransferID back to the receiver
				newTransferID := gen_TransferID()

				utils.FT_Map[newTransferID] = &utils.FT_Transfer{
					ReceiverConn:     conn,
					TransferID:       newTransferID,
					ReceiverApproved: false,
				}

				cmd := FT_Command{
					SentBy:  "server",
					Command: "gen_TransferID",
					Payload: newTransferID,
				}
				SendCommandToConn(&cmd, conn)

			case "start_Transfer":
				if !verify_TransferID(command.TransferID) {
					SendCommandToConn(&ErrInvalidTransferID, conn)
					conn.Close()
					return nil
				}

				cmd := FT_Command{
					SentBy:  "server",
					Command: "start_transfer",
					Payload: "",
				}
				SendCommandToConn(&cmd, utils.FT_Map[command.TransferID].SenderConn)

			case "end_Transfer":
				continue

			}

		} else if command.SentBy == "sender" {
			switch command.Command {
			case "sender_metadata":
				// Handle
				if !verify_TransferID(command.TransferID) {
					SendCommandToConn(&ErrInvalidTransferID, conn)
					conn.Close()
					return nil
				}

				MD_Payload, ok := command.Payload.(FT_MetaData)
				if !ok {
					SendCommandToConn(&ErrInvalidFormat, conn)
					conn.Close()
					return nil
				}

				utils.FT_Map[command.TransferID].FileByteSize = MD_Payload.FileByteSize
				utils.FT_Map[command.TransferID].Filename = MD_Payload.Filename
				utils.FT_Map[command.TransferID].SenderConn = conn

				cmd := FT_Command{
					SentBy:  "server",
					Command: "sender_metadata",
					Payload: &FT_MetaData{
						TransferID:   MD_Payload.TransferID,
						Filename:     MD_Payload.Filename,
						FileByteSize: MD_Payload.FileByteSize,
					},
				}
				SendCommandToConn(&cmd, utils.FT_Map[MD_Payload.TransferID].ReceiverConn)

			case "data_packet":
				if !verify_TransferID(command.TransferID) {
					SendCommandToConn(&ErrInvalidTransferID, conn)
					conn.Close()
					return nil
				}

				dataPacket, ok := command.Payload.(FT_Packet)
				if !ok {
					SendCommandToConn(&ErrInvalidFormat, conn)
					conn.Close()
					return nil
				}

				// Payload is expected to be of type FT_Packet
				cmd := FT_Command{
					SentBy:  "server",
					Command: "data_packet",
					Payload: dataPacket,
				}
				SendCommandToConn(&cmd, utils.FT_Map[command.TransferID].ReceiverConn)

			case "end_Transfer":
				if !verify_TransferID(command.TransferID) {
					SendCommandToConn(&ErrInvalidTransferID, conn)
					conn.Close()
					return nil
				}

				cmd := FT_Command{
					SentBy:  "server",
					Command: "end_transfer",
					Payload: "",
				}
				SendCommandToConn(&cmd, utils.FT_Map[command.TransferID].ReceiverConn)
			}
		}
	}
}
