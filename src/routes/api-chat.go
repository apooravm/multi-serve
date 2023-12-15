package routes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func Chat(c echo.Context) error {
	conn, err := utils.ConnUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	socketInitFlag := false
	greetingFlag := false

	client_id := strconv.Itoa(utils.Id_Gen.GenerateNewID())
	var newClient utils.Client
	newClient.Username = "undefined"

	for {
		var clientMessage utils.Message
		if err := conn.ReadJSON(&clientMessage); err != nil {
			ClientDisconnect(conn, client_id, &newClient)
			return nil
		}

		if clientMessage.Password != utils.CHAT_PASS {
			conn.Close()
			return nil
		}

		if !greetingFlag {
			ServerMessageToClient("Welcome to the Server!", conn)
			greetingFlag = true
		}

		// Init handshake
		// setup Config-Type = config-username
		if clientMessage.Direction == utils.C2S {
			config_split := strings.Split(clientMessage.Config, "-")
			if config_split[0] == "config" {
				if config_split[1] == "username" {
					newClient = utils.Client{
						Id:       client_id,
						Username: clientMessage.Sender,
						Conn:     conn,
					}

					utils.ChatClientsMap.AddClient(client_id, &newClient)

					// List All the client, server-side only for now
				} else if config_split[1] == "list" {
					BroadcastServerMessageAll("Current Online => " + utils.ChatClientsMap.GetClientsStr())

				} else if config_split[1] == "close" {
					// Socket disconnection
					ClientDisconnect(conn, client_id, &newClient)
					return nil
				}
			}
		} else if clientMessage.Direction == utils.C2A {
			utils.LogData(fmt.Sprintf("%v: %v", clientMessage.Sender, clientMessage.Content), utils.CHAT_LOG)
			BroadcastClientMessageAll(clientMessage, client_id)
		}

		if !socketInitFlag {
			message := clientMessage.Sender + " joined!"
			utils.LogData(message, utils.CHAT_LOG)
			BroadcastServerMessageAll(message)
			socketInitFlag = true
		}
	}
}

// Server to a single client
// Main method called by others
func SendMessageToClient(message utils.Message, conn *websocket.Conn) error {
	if err := conn.WriteJSON(message); err != nil {
		utils.LogData(fmt.Sprintln("Error sending message to client:", err), utils.CHAT_DEBUG)
		return &utils.ServerError{
			Err:    err,
			Code:   utils.SERVER_ERR,
			Simple: "Error sending message to client",
		}
	}
	return nil
}

// Simple wrapper to directly pass in strings and send to a conn
func ServerMessageToClient(payload string, conn *websocket.Conn) {
	message := utils.Message{
		Sender:    "Server",
		Direction: utils.S2C,
		Config:    "",
		Content:   payload,
		Password:  "",
	}
	if err := SendMessageToClient(message, conn); err != nil {
		utils.LogData(err.Error(), utils.CHAT_DEBUG)
		return
	}
}

// Main chat broadcast
// General chat
// The client who sent the message is skipped
func BroadcastClientMessageAll(message utils.Message, client_id string) {
	for id, client := range utils.ChatClientsMap.Clients {
		if id != client_id {
			if err := SendMessageToClient(message, client.Conn); err != nil {
				return
			}
		}
	}
}

// For Server notification broadcasting
// When user joins/disconnects etc
func BroadcastServerMessageAll(payload string) {
	message := utils.Message{
		Sender:    "Server",
		Direction: utils.S2C,
		Config:    "",
		Content:   payload,
		Password:  "",
	}
	for _, client := range utils.ChatClientsMap.Clients {
		if err := SendMessageToClient(message, client.Conn); err != nil {
			return
		}
	}
}

// Handle Client socket disconnection
// Graceful handling prevents error logs
func ClientDisconnect(conn *websocket.Conn, client_id string, client *utils.Client) {
	conn.Close()
	// Handle disconnection or error here
	// Delete client from the map
	utils.ChatClientsMap.DeleteClient(client_id)
	message := client.Username + " left!"
	utils.LogData(message, utils.CHAT_LOG)
	BroadcastServerMessageAll(message)
}
