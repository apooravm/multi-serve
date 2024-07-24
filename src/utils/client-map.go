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

// Hashmap of a client against their id.
type ClientsMap[T any] struct {
	Clients map[string]T
	mu      sync.RWMutex
}

func NewClientsMap[T any]() *ClientsMap[T] {
	return &ClientsMap[T]{
		Clients: make(map[string]T),
	}
}

func (c *ClientsMap[T]) AddClient(clientID string, client *T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Clients[clientID] = *client
}

func (c *ClientsMap[T]) DeleteClient(clientID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Clients, clientID)
}

func (c *ClientsMap[T]) GetClient(clientID string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	client, ok := c.Clients[clientID]
	return client, ok
}

func (c *ClientsMap[T]) UpdateClient(clientID string, updateFunc func(client T) T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if client, exists := c.Clients[clientID]; exists {
		c.Clients[clientID] = updateFunc(client)
	}
}

// Returns a string of all the online clients
func GetClientsStr(c *ClientsMap[Client]) string {
	usernameArr := []string{}
	for _, client := range c.Clients {
		usernameArr = append(usernameArr, client.Username)
	}
	return strings.Join(usernameArr, " | ")
}
