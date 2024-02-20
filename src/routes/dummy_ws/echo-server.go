package dummy_ws

import (
	"fmt"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func EchoDummyWS(c echo.Context) error {
	conn, err := utils.ConnUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	// hasBroadcastJoinedMsg := false
	// greetingFlag := false

	// client_id := strconv.Itoa(utils.Id_Gen.GenerateNewID())
	var newClient utils.Client
	newClient.Username = "undefined"

	for {
		var clientData interface{}
		if err := conn.ReadJSON(&clientData); err != nil {
			clientDisconnect(conn)
			return nil
		}

	}

	return nil
}

// Handle Client socket disconnection
// Graceful handling prevents error logs
func clientDisconnect(conn *websocket.Conn) {
	fmt.Println("Conn Left")
	conn.Close()
	// Handle disconnection or error here
	// // Delete client from the map
	// utils.ChatClientsMap.DeleteClient(client_id)
	// message := client.Username + " left!"
	// utils.LogData(message, utils.CHAT_LOG)
	// BroadcastServerMessageAll(message)
}
