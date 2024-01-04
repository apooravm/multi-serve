package routes

import (
	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func MiscGroup(group *echo.Group) {
	group.GET("/stream", StreamVideoFile)
	group.GET("/greet", RandomGreeting)
	group.GET("/echo", EchoBackQuery)
	group.GET("/echo/json", EchoBackBody)
}

func EchoBackQuery(c echo.Context) error {
	data := c.QueryParam("data")

	return c.JSON(200, data)
}

func EchoBackBody(c echo.Context) error {
	var bodyContent interface{}

	if err := c.Bind(&bodyContent); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error binding:"+err.Error()))
	}

	return c.JSON(200, bodyContent)
}

func StreamVideoFile(c echo.Context) error {
	return c.JSON(200, &utils.SuccessMessage{Message: "Under Construction"})
}

func RandomGreeting(c echo.Context) error {
	return c.String(200, "Have a good day")
}
