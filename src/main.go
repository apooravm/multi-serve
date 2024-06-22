package main

import (
	"log"
	"os"

	"github.com/apooravm/multi-serve/src/routes"
	"github.com/apooravm/multi-serve/src/utils"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "dev" {
			if err := godotenv.Load(); err != nil {
				log.Println("Error loading .env file")
			}
		}
	}

	utils.InitGlobalVars()
	utils.InitDirs()
	utils.S3_ObjectInfoArr()
	utils.InitFiles()

	// Download files; resume and such
	if err := utils.S3_DownloadFiles(); err != nil {
		utils.LogData("main.go err_id:001 | error downloading S3 files", err.Error())

	} else {
		utils.LogData("S3 Files downloaded successfully 🎉")
	}

	// Download notes data
	if err := utils.DownloadAndWriteNoteData(); err != nil {
		utils.LogData("main.go err_id:002 | error downloading note files", err.Error())

	} else {
		utils.LogData("S3 Notes downloaded successfully 🙌")
	}

	PORT := utils.PORT
	utils.LogData("Live on PORT", PORT, "🔥")

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
				utils.LogData("main.go err_id:003 |", err.Error())
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Static("/", "public")
	e.GET("/help", handleNotesRes)

	DefaultGroup(e.Group(""))
	e.Logger.Fatal(e.Start(":" + PORT))
}

func handleNotesRes(c echo.Context) error {
	return c.File("./helper.txt")
}

func DefaultGroup(group *echo.Group) {
	routes.ApiGroup(group.Group("/api"))
}
