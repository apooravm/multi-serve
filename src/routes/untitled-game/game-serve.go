package untitledgame

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type PlayerPayload struct {
	Px   uint16 // PosX
	Py   uint16 // PosY
	Name string // Username
	Dir  string // Direction
	Cfg  uint8  // Config
	Pwd  string // Password
	Id   string // Id
}

func UntitledGameSocket(c echo.Context) error {
	utils.ConnUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := utils.ConnUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	// hasBroadcastJoinedMsg := false
	// greetingFlag := false

	client_id := strconv.Itoa(utils.Id_Gen.GenerateNewID())
	var player utils.PlayerClient
	player.Username = "undefined"
	player.Conn = conn
	player.Id = client_id

	log.Println("New User here", client_id)
	log.Println(utils.GamePlayersMap)

	for {
		var clientMessage PlayerPayload
		if err := conn.ReadJSON(&clientMessage); err != nil {
			ClientDisconnect(&player)
			return nil
		}

		if clientMessage.Pwd != utils.CHAT_PASS {
			// conn.Close()
			ClientDisconnect(&player)
			return nil
		}

		// if !greetingFlag {
		// 	ServerMessageToClient("Welcome to the Server!", conn)
		// 	greetingFlag = true
		// }

		// Init handshake
		// setup Config-Type = config-username
		if clientMessage.Dir == utils.C2S {
			// config_split := strings.Split(clientMessage.Config, "-")
			// if clientMessage.Config == 0 {
			// 	ClientDisconnect(&player)
			// }

			// Dropping strings for Config. Think string splitting and checkup would take too long
			// 1 -> Initial handshake
			// 2 -> Voluntary disconnect
			// 3 -> Player position update
			switch clientMessage.Cfg {
			case 1:
				player = utils.PlayerClient{
					Id:       client_id,
					Username: clientMessage.Name,
					Conn:     conn,
					PosX:     10,
					PosY:     10,
				}

				utils.GamePlayersMap.AddClient(client_id, &player)

			case 2:
				log.Println("Voluntary disconnect")
				ClientDisconnect(&player)
				return nil

			case 3:
				utils.GamePlayersMap.UpdateClient(client_id, func(client *utils.PlayerClient) *utils.PlayerClient {
					client.PosX = clientMessage.Px
					client.PosY = clientMessage.Py
					return client
				})
				BroadcastPlayerPaylaodAll(client_id, player.Username, player.PosX, player.PosY, 3)
			}

		} else if clientMessage.Dir == utils.C2A {
			// utils.LogDataToPath(utils.CHAT_LOG, fmt.Sprintf("%v: %v", clientMessage.Username, clientMessage.Content))
			// BroadcastClientMessageAll(clientMessage, client_id)
		}

		// if !hasBroadcastJoinedMsg {
		// 	message := clientMessage.Username + " joined!"
		// 	utils.LogDataToPath(utils.CHAT_LOG, message)
		// 	BroadcastServerMessageAll(message)
		// 	hasBroadcastJoinedMsg = true
		// }
	}
}

// Server to a single client
// Main method called by others
func SendMessageToClient(message *PlayerPayload, conn *websocket.Conn) error {
	if err := conn.WriteJSON(message); err != nil {
		utils.LogDataToPath(utils.CHAT_DEBUG, fmt.Sprintln("game-serve.go err_id:001 | error sending payload to client:", err))
		return &utils.ServerError{
			Err:    err,
			Code:   utils.SERVER_ERR,
			Simple: "Error sending message to client",
		}
	}
	return nil
}

// Simple wrapper to directly pass in strings and send to a conn
// func ServerMessageToClient(payload string, conn *websocket.Conn) {
// 	message := utils.Message{
// 		Sender:    "Server",
// 		Direction: utils.S2C,
// 		Config:    "",
// 		Content:   payload,
// 		Password:  "",
// 	}
// 	if err := SendMessageToClient(message, conn); err != nil {
// 		utils.LogDataToPath(utils.CHAT_DEBUG, "api-chat.go err_id:002 | error sending server message to client", err.Error())
// 		return
// 	}
// }

// Main chat broadcast
// General chat
// The client who sent the message is skipped
// func BroadcastClientMessageAll(message utils.Message, client_id string) {
// 	for id, client := range utils.GamePlayersMap.Clients {
// 		if id != client_id {
// 			if err := SendMessageToClient(message, client.Conn); err != nil {
// 				return
// 			}
// 		}
// 	}
// }

// For Server notification broadcasting
// When user joins/disconnects etc
func BroadcastPlayerPaylaodAll(clientID string, username string, posX uint16, posY uint16, config uint8) {
	message := PlayerPayload{
		Px:   posX,
		Py:   posY,
		Name: username,
		Dir:  utils.S2C,
		Cfg:  config,
		Pwd:  "",
		Id:   clientID,
	}
	for _, client := range utils.GamePlayersMap.Clients {
		if err := SendMessageToClient(&message, client.Conn); err != nil {
			return
		}
	}
}

// Handle Client socket disconnection
// Graceful handling prevents error logs
func ClientDisconnect(client *utils.PlayerClient) {
	client.Conn.Close()
	// Handle disconnection or error here
	// Delete client from the map
	utils.GamePlayersMap.DeleteClient(client.Id)
	// message := client.Username + " left!"
	// utils.LogDataToPath(utils.CHAT_LOG, message)
	// BroadcastServerMessageAll(message)
}
