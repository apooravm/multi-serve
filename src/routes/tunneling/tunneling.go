package tunneling

import (
	"fmt"
	"io"
	"net/http"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func Tunneling(c echo.Context) error {
	id := c.Request().URL.Query().Get("id")

	fmt.Println("ID", id)

	if id == "" {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("No id found."))
	}

	byteData, err := io.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Println(err.Error())
		utils.LogData("tunneling.go E:Failed to read body data.")
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Failed to read body contents."))
	}

	fmt.Println(string(string(byteData)))

	utils.ConnUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := utils.ConnUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		utils.LogData("tunneling.go E:Failed to upgrade websocket connection.")
		fmt.Println(err.Error())
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Failed to upgrade ws conn."))
	}

	defer conn.Close()

	return nil
}
