package routes

import (
	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func MiscGroup(group *echo.Group) {
	group.GET("/stream", StreamVideoFile)
}

func StreamVideoFile(c echo.Context) error {
	return c.JSON(200, &utils.SuccessMessage{Message: "Under Construction"})
}
