package utils

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Id       string
	Username string
	Conn     *websocket.Conn
}

type Message struct {
	Sender    string
	Direction string
	Config    string
	Content   string
	Password  string
	Timestamp string
}

/*
interface Message {
    Id: number
    Sender: string;
    Direction: string;
    Config: string;
    Content: string;
    Password: string;
    Timestamp: string;
  }

*/

type ClientsMap struct {
	Clients map[string]Client
	mu      sync.RWMutex
}

const (
	C2A = "client-to-all"
	C2S = "client-to-server"
	S2C = "server-to-client" // Server to single client broadcast
	S2A = "server-to-all"    // Global client broadcast
)

const (
	SERVER_ERR = 500
	CLIENT_ERR = 400
)

// Custom Server Error
type ServerError struct {
	Err    error
	Code   int
	Simple string
}

func (ce *ServerError) Error() string {
	return fmt.Sprintf("%v", ce.Simple)
}

func NewClientsMap() *ClientsMap {
	return &ClientsMap{
		Clients: make(map[string]Client),
	}
}

func (c *ClientsMap) AddClient(clientID string, client *Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Clients[clientID] = *client
}

func (c *ClientsMap) DeleteClient(clientID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Clients, clientID)
}

func (c *ClientsMap) GetClient(clientID string) (Client, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	client, ok := c.Clients[clientID]
	return client, ok
}

// Returns a string of all the online clients
func (c *ClientsMap) GetClientsStr() string {
	usernameArr := []string{}
	for _, client := range c.Clients {
		usernameArr = append(usernameArr, client.Username)
	}
	return strings.Join(usernameArr, " | ")
}
