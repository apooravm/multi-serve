package routes

import (
	"fmt"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func JournalLoggerGroup(group *echo.Group) {
	group.GET("/", GetJournalLogs)
	group.POST("/log", PostJournalLogEntry)
}

func GetJournalLogs(c echo.Context) error {
	username := c.Param("username")
	password := c.Param("password")

	fmt.Println(username, password)
	return c.JSON(200, &utils.SuccessMessage{Message: "Under Construction"})
}

func PostJournalLogEntry(c echo.Context) error {
	return c.JSON(200, &utils.SuccessMessage{Message: "Under Construction"})
}
