package main

import (
	"fmt"
	"log"
	"os"

	"github.com/apooravm/multi-serve/src/routes"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file")
	}
	PORT := os.Getenv("PORT")
	e := echo.New()
	e.Use(middleware.CORS())

	e.Static("/", "public")

	e.GET("/api/signup", routes.SignupHandler)
	e.GET("/api/login", routes.LoginHandler)
	e.GET("/api/resume", routes.GetResume)
	e.GET("/api/cronping", routes.CronPing)

	fmt.Printf("Live on %v", PORT)
	e.Logger.Fatal(e.Start(":" + PORT))
}
