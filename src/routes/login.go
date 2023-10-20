package routes

import "github.com/labstack/echo/v4"

func LoginHandler(c echo.Context) error {
	return c.String(200, "login working")
}
