package utils

import "github.com/gorilla/websocket"

var (
	ChatClientsMap               = NewClientsMap[Client]()
	Id_Gen         *id_Generator = &id_Generator{
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
