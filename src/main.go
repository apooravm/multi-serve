package main

import (
	"fmt"
	"log"

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

	utils.InitGlobalVars()
	utils.InitDirs()
	utils.InitVars()

	// Download files; resume and such
	if err := utils.S3_DownloadFiles(); err != nil {
		utils.LogData(err.Error(), utils.SERVER_LOG_PATH)

	} else {
		utils.LogData("S3 Files downloaded successfully", utils.SERVER_LOG_PATH)
	}

	// Download notes data
	if err := utils.DownloadAndWriteNoteData(); err != nil {
		utils.LogData(err.Error(), utils.SERVER_LOG_PATH)

	} else {
		utils.LogData("S3 Notes downloaded successfully", utils.SERVER_LOG_PATH)
	}

	PORT := utils.PORT

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
			data := utils.Log{
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
			if err := utils.AppendLogToFile(&data, utils.REQUEST_LOG_PATH); err != nil {
				fmt.Println("main.go ln:69 |", err)
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
