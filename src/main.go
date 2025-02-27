package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/apooravm/multi-serve/src/routes"
	"github.com/apooravm/multi-serve/src/utils"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	Host struct {
		Echo *echo.Echo
	}
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

	hosts := map[string]*Host{}

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

	blog := echo.New()
	blog.Use(middleware.Logger())
	blog.Use(middleware.Recover())

	blog_endpoint := fmt.Sprintf("blog.localhost:%s", PORT)
	hosts[blog_endpoint] = &Host{blog}

	blog.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Blog")
	})

	api := echo.New()
	api.Use(middleware.CORS())
	api.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
	api.Use(middleware.Recover())
	api.Static("/", "public")
	api_endpoint := fmt.Sprintf("localhost:%s", PORT)
	hosts[api_endpoint] = &Host{api}

	api.GET("/help", handleNotesRes)
	DefaultGroup(api.Group(""))

	server := echo.New()
	server.Any("/*", func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		host := hosts[req.Host]
		var err error

		if host == nil {
			err = echo.ErrNotFound

		} else {
			host.Echo.ServeHTTP(res, req)
		}

		return err
	})

	server.Logger.Fatal(server.Start(":" + PORT))
}

func handleNotesRes(c echo.Context) error {
	return c.File("./helper.txt")
}

func DefaultGroup(group *echo.Group) {
	routes.ApiGroup(group.Group("/api"))
}
