package routes

import (
	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func MiscGroup(group *echo.Group) {
	group.GET("/stream", StreamVideoFile)
	group.GET("/greet", RandomGreeting)
}

func StreamVideoFile(c echo.Context) error {
	return c.JSON(200, &utils.SuccessMessage{Message: "Under Construction"})
}

func RandomGreeting(c echo.Context) error {
	return c.String(200, "Have a good day")
}
