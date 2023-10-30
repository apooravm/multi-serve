package routes

import (
	"log"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var clients = make(map[string]*websocket.Conn)
var clientsMu sync.RWMutex

var (
	upgrader  = websocket.Upgrader{}
	unique_id = 1
)

type ClientMessage struct {
	Username   string `json:"username"`
	ClientID   string `json:"clientID"`
	Message    string `json:"message"`
	ClientRoom int    `json:"clientRoom"`
}

func Chat(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// onConnection to a client
	clientID := strconv.Itoa(unique_id) // Replace with a unique identifier
	// Add the client to the map of connected clients
	clientsMu.Lock()
	clients[clientID] = ws
	clientsMu.Unlock()

	if err = ws.WriteJSON(ClientMessage{
		Username:   "server",
		ClientID:   clientID,
		Message:    "Welcome to the Server",
		ClientRoom: 0,
	}); err != nil {
		c.Logger().Error(err)
	}

	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		clientsMu.Lock()
		broadcastMessageToAll(messageType, &p, clientID)
		clientsMu.Unlock()

	}

	clientsMu.Lock()
	delete(clients, clientID)
	clientsMu.Unlock()
	return nil
}

func broadcastMessageToAll(messageType int, data *[]byte, clientID string) error {
	for id, client := range clients {
		if id != clientID {
			if err := client.WriteMessage(messageType, *data); err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}
