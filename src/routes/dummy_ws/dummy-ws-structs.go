package dummy_ws

import (
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Id       string
	Username string
	Conn     *websocket.Conn
}

type ClientsMap struct {
	Clients map[string]Client
	mu      sync.RWMutex
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
