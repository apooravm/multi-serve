package dummy_ws

import (
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	UserMap            *ClientsMap   = NewClientsMap()
	SubscribedUsersMap *ClientsMap   = NewClientsMap()
	Id_Gen             *id_Generator = &id_Generator{
		start_ID: 0,
	}
	ConnUpgrader = websocket.Upgrader{}
)

type id_Generator struct {
	start_ID int
}

func (idGen *id_Generator) GenerateNewID() int {
	ret_id := idGen.start_ID
	idGen.start_ID += 1

	return ret_id
}

type Client struct {
	Id   string
	Conn *websocket.Conn
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
		usernameArr = append(usernameArr, client.Id)
	}
	return strings.Join(usernameArr, " | ")
}
