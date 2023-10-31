package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/apooravm/multi-serve/src/routes"
	"github.com/apooravm/multi-serve/src/utils"

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
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		BeforeNextFunc: func(c echo.Context) {
			c.Set("customValueOnRequest", 42)
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			_ = c.Get("customValueOnRequest")
			file, err := os.OpenFile("./data/logs.json", os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				fmt.Println("Failed to open log file")
				return err
			}

			defer file.Close()

			var logs []utils.Log

			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&logs); err != nil {
				if err != io.EOF {
					fmt.Println(err)
					return nil
				}
			}

			newLog := utils.Log{
				ContentLength: v.ContentLength,
				Error:         v.Error,
				Host:          v.Host,
				Latency:       v.Latency,
				RemoteIP:      v.RemoteIP,
				ResponseSize:  v.ResponseSize,
				Time:          v.StartTime,
				Status:        v.Status,
				URI:           v.URI,
				Protocol:      v.Protocol,
			}

			logs = append(logs, newLog)

			updatedJSON, err := json.MarshalIndent(logs, "", "    ")
			if err != nil {
				fmt.Println(err)
				return nil
			}

			if _, err := file.Seek(0, 0); err != nil {
				fmt.Println(err)
				return nil
			}
			if err := file.Truncate(0); err != nil {
				fmt.Println(err)
				return nil
			}
			_, err = file.Write(updatedJSON)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Static("/", "public")

	DefaultGroup(e.Group(""))
	fmt.Printf("Live on %v", PORT)
	e.Logger.Fatal(e.Start(":" + PORT))
}

func DefaultGroup(group *echo.Group) {
	routes.ApiGroup(group.Group("/api"))
}
