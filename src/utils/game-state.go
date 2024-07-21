package utils

import "github.com/gorilla/websocket"

var (
	GamePlayersMap *ClientsMap[PlayerClient] = NewClientsMap[PlayerClient]()
)

type PlayerClient struct {
	Id       string
	Username string
	Conn     *websocket.Conn
	PosX     uint16
	PosY     uint16
}
