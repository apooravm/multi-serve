package tunneling

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

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
	TunnelHostMap *utils.ClientsMap[TunnelHost] = utils.NewClientsMap[TunnelHost]()
)

// Differentiating between CLI and browser
// If tunneling/host -> Host CLI
// If tunneling/<CODE>/ -> client browser
func TunnelingGroup(group *echo.Group) {
	group.GET("/host", HandleHostTunnel)
	// redirect /:code to /:code/
	group.GET("/:code", func(c echo.Context) error {
		code := c.Param("code")
		return c.Redirect(http.StatusPermanentRedirect, "/api/tunnel/"+code+"/")
	})
	group.Any("/:code/*", HandleClientTunnel)
}

// Creates a binary packet with the given components. Byte order big endian.
// First part should be the VERSION.
// Second the MESSAGE_CODE.
// Third the DATA (optional).
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
func HandleHostTunnel(c echo.Context) error {
	// TODO: Authorize conn ...
	// TODO: Make it work with > uint8 unique_code

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

	var code uint8 = utils.Tunneling_code_generator.NewCode()
	pkt, _ := CreateBinaryPacket(byte(1), byte(1), code)

	// send the newly generated code back to CLI
	if err := conn.WriteMessage(websocket.BinaryMessage, pkt); err != nil {
		utils.LogData("E:tunneling.go Could not write code to host socket.")
		conn.Close()

		return nil
	}

	// generate code
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

	}
}

type TunnelRequestPayload struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

type TunnelResponsePayload struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

func HandleClientTunnel(c echo.Context) error {
	incomingCode := c.Param("code")
	routepath := c.Param("*")

	fmt.Println("Route path:", routepath)

	// Check if transfer exists
	targetTunnel, exists := TunnelHostMap.GetClient(incomingCode)
	if !exists {
		// return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Unknown code. Transfer not found."))
		return c.JSON(echo.ErrUnauthorized.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Unknown code. Transfer not found.",
		})
	}

	reqBodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Failed to read body",
		})
	}

	reqHeaders := map[string]string{}
	for k, v := range c.Request().Header {
		reqHeaders[k] = strings.Join(v, ",")
	}

	reqPayload := TunnelRequestPayload{
		Method:  c.Request().Method,
		Body:    reqBodyBytes,
		Headers: reqHeaders,
		Path:    routepath,
	}

	var requestBuf bytes.Buffer
	enc := gob.NewEncoder(&requestBuf)
	if err = enc.Encode(reqPayload); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "E: Serializing to request buffer.",
		})
	}

	// fmt.Println("Tunnel found")
	// fmt.Println(targetTunnel.TunnelName)

	// Request data here, message code = 3
	pkt, _ := CreateBinaryPacket(byte(1), byte(2), requestBuf.Bytes())
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

	switch message[1] {
	case byte(3):
		// html_data = string(message[2:])
		var responsePayloadBuf TunnelResponsePayload
		dec := gob.NewDecoder(bytes.NewReader(message[2:]))
		if err := dec.Decode(&responsePayloadBuf); err != nil {
			return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
				Code:    echo.ErrInternalServerError.Code,
				Message: "E: Decoding serialized response buffer.",
			})
		}

		// write headers, statuscode and headers to res
		for k, v := range responsePayloadBuf.Headers {
			c.Response().Header().Set(k, v)
		}

		c.Response().WriteHeader(responsePayloadBuf.StatusCode)
		_, err := c.Response().Write(responsePayloadBuf.Body)
		if err != nil {
			return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
				Code:    echo.ErrInternalServerError.Code,
				Message: "E: Writing body bytes to response.",
			})
		}
	}

	return err
}
