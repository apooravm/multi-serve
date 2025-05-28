package tunneling

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type TunnelHost struct {
	Code            uint8
	TunnelName      string
	TunnelConn      *websocket.Conn
	TunnelClosed    bool
	stopCh          chan struct{}
	TransferStopped bool
}

var (
	TunnelHostMap  *utils.ClientsMap[TunnelHost] = utils.NewClientsMap[TunnelHost]()
	code_generator utils.Code_Generator          = utils.Code_Generator{
		Start_ID: 1,
	}
)

// Differentiating between CLI and browser
// If tunneling/host -> Host CLI
// If tunneling/<CODE>/ -> client browser
func TunnelingGroup(group *echo.Group) {
	group.GET("/host", HandlHostTunnel)
	// redirect /:code to /:code/
	group.GET("/:code", func(c echo.Context) error {
		code := c.Param("code")
		return c.Redirect(http.StatusPermanentRedirect, "/api/tunnel/"+code+"/")
	})
	group.GET("/:code/*", HandleClientTunnel)
}

// Creates a binary packet with the given components. Byte order big endian
// First part should be the version
// Second the message code
// Third the data if available
func CreateBinaryPacket(parts ...any) ([]byte, error) {
	responseBfr := new(bytes.Buffer)
	for _, part := range parts {
		if err := binary.Write(responseBfr, binary.BigEndian, part); err != nil {
			return nil, err
		}
	}

	return responseBfr.Bytes(), nil
}

// TUNNELING PROTOCOL MESSAGE CODES
// 1 - Requesting unique code, data = nothing
// Server generates code
// 2 - Sending unique code, data = "code"
// User on serverside goes to certain path (unclear on how this is going to work for now)
// 3 - Requesting data with url-route, data = "/path/to/whatever"
// Client combines the path with the base endpoint and sends a request and gets back data
// 4 - Sending back the data corresponding to the path, data = "..."

// PACKET FRAME
// [VERSION][MESSAGE_CODE][DATA (opt)]

// CLI sends ws req here
// Check for validity, etc
// Send back unique code
// Register same unique code as a route, ie; tunneling/:1234/...
func Tunneling(c echo.Context) error {
	params := c.QueryParams()
	intentArgs := params["intent"]

	if len(intentArgs) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("No intent found. Must be send/receive."))
	}

	// var initialError error = nil

	// Client part should end in it returning HTML. No need to keep it alive (For now).

	return nil
}

func HandlHostTunnel(c echo.Context) error {
	// TODO: Authorize conn ...
	// TODO: Generate unique code
	// TODO: Make it work with > uint8 unique_code

	var code uint8 = 255
	// buf := new(bytes.Buffer)
	// // TODO: Handle error here
	// _ = binary.Write(buf, binary.BigEndian, int32(code))
	//
	// fmt.Println("writing code")
	//
	// conn.WriteMessage(websocket.BinaryMessage, []byte("1234"))
	utils.ConnUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := utils.ConnUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("WS_ERR:", err.Error())
		return err
	}

	defer conn.Close()

	fmt.Println("Host here")
	// generate code
	code = 255
	var newTunnelingHost TunnelHost
	newTunnelingHost.Code = code
	newTunnelingHost.TransferStopped = false
	newTunnelingHost.TunnelClosed = false
	// newTunnelingHost.TunnelConn =
	newTunnelingHost.TunnelName = "test tunnel"
	newTunnelingHost.TunnelConn = conn
	newTunnelingHost.Code = code

	fmt.Printf("Host saved with code %s\n", strconv.Itoa(int(code)))
	TunnelHostMap.AddClient(strconv.Itoa(int(code)), &newTunnelingHost)

	// Keeps conn alive
	for {
		// fmt.Println("tf??")
		// var conn_message []byte
		// _, conn_message, err := conn.ReadMessage()
		// if err != nil {
		// 	conn.Close()
		// 	return nil
		// }
		//
		// switch conn_message[1] {
		// case 1:
		// 	var currentTunnelHost TunnelHost
		// 	currentTunnelHost.TunnelConn = conn
		// 	currentTunnelHost.TunnelName = "cool_tunnel"
		// 	currentTunnelHost.Code = code
		//
		// 	TunnelHostMap.AddClient(strconv.Itoa(int(code)), &currentTunnelHost)
		//
		// 	fmt.Println(TunnelHostMap)
		//
		// 	pkt, err := CreateBinaryPacket(byte(1), byte(2), code)
		// 	if err != nil {
		// 		log.Println("E:tunneling.go Could not create binary packet", err.Error())
		// 		conn.Close()
		// 		return nil
		// 	}
		//
		// 	if err = conn.WriteMessage(websocket.BinaryMessage, pkt); err != nil {
		// 		log.Println("E:tunneling.go Could not write to socket", err.Error())
		// 		conn.Close()
		// 		return nil
		// 	}
		//
		// default:
		// 	fmt.Println("🤷‍♀️")
		// 	conn.Close()
		// 	return nil
		// }
	}
}

func HandleClientTunnel(c echo.Context) error {
	incomingCode := c.Param("code")
	routepath := c.Param("*")

	// Check if transfer exists
	targetTunnel, exists := TunnelHostMap.GetClient(incomingCode)
	if !exists {
		// return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Unknown code. Transfer not found."))
		return c.JSON(echo.ErrUnauthorized.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Unknown code. Transfer not found.",
		})
	}

	// fmt.Println("Tunnel found")
	// fmt.Println(targetTunnel.TunnelName)

	// Request data here, message code = 3
	pkt, _ := CreateBinaryPacket(byte(1), byte(3), []byte(routepath))
	if err := targetTunnel.TunnelConn.WriteMessage(websocket.BinaryMessage, pkt); err != nil {
		log.Println("E:tunneling.go Could not write to socket", err.Error())
		targetTunnel.TunnelConn.Close()
		targetTunnel.TunnelClosed = true
		return nil
	}

	// fmt.Println("Sending message to client")
	// fmt.Println(pkt)
	// Raw html sent for now, need to serialize it into frames
	messageType, message, err := targetTunnel.TunnelConn.ReadMessage()
	if err != nil {
		fmt.Println("E:error reading from host socket")
		return nil
	}

	if messageType != websocket.BinaryMessage {
		fmt.Println("E:Invalid message type from sender")
		return nil
	}

	return c.HTML(http.StatusOK, string(message))
}
