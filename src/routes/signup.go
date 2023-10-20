package routes

import (
	"github.com/labstack/echo/v4"
)

func SignupHandler(c echo.Context) error {
	return c.String(200, "signup working")
}
