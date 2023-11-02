package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

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
}

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

var (
	CHAT_DEBUG = os.Getenv("CHAT_DEBUG")
	CHAT_LOG   = os.Getenv("CHAT_LOG")
	CHAT_PASS  = os.Getenv("CHAT_PASS")
)

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

func LogData(data string, logFilePath string) {
	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		moreErr := ServerError{
			Err:    err,
			Code:   SERVER_ERR,
			Simple: "Error opening the log file",
		}
		fmt.Println(moreErr.Error())
	}
	defer file.Close()

	currentTime := time.Now()
	timeString := currentTime.Format("2006-01-02 15:04:05")
	data = timeString + " " + data + "\n"

	_, err = file.WriteString(data)
	if err != nil {
		moreErr := ServerError{
			Err:    err,
			Code:   SERVER_ERR,
			Simple: "Error logging",
		}
		fmt.Println(moreErr.Error())
	}
}
