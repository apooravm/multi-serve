package routes

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func ApiGroup(group *echo.Group) {
	group.GET("/resume", GetResume)
	group.GET("/cronping", CronPing)
	group.GET("/logs", GetLoggedData)

	group.GET("/chat", Chat)
	NotesGroup(group.Group("/notes"))
}

// Initially download the file in ./data/files/resume.pdf
// If file exists at path then return the file
// Else download to the location
// Reduction in S3 fetching cost
func GetResume(c echo.Context) error {
	resumeFilePath := os.Getenv("LOCAL_RESUME_PATH")

	// Downloading the file from Bucket
	if !utils.FileExists(resumeFilePath) {
		file, err := utils.DownloadFile(os.Getenv("BUCKET_NAME"),
			os.Getenv("OBJ_RESUME_KEY"),
			os.Getenv("BUCKET_REGION"))

		if err != nil {
			errMsg := utils.ErrorMessage{
				Code:    echo.ErrInternalServerError.Code,
				Message: "Server Error, Error fetching data from S3 Bucket",
			}
			fmt.Println(errMsg)
			fmt.Println("\nERR\n", err.Error())
			return c.JSON(echo.ErrInternalServerError.Code, &errMsg)
		}

		path_split := strings.Split(resumeFilePath, "/")
		mkdirPath := strings.Join(path_split[:len(path_split)-1], "/")
		err = os.MkdirAll(mkdirPath, os.ModePerm)
		if err != nil {
			panic(err)
		}

		// Writing the file to dir
		if err := os.WriteFile(resumeFilePath, file, 0644); err != nil {
			return c.JSON(echo.ErrInternalServerError.Code, utils.ErrorMessage{
				Code:    echo.ErrInternalServerError.Code,
				Message: "Serve Error, Error writing to local file. " + err.Error(),
			})
		}
	}
	return c.File(resumeFilePath)
}

func CronPing(c echo.Context) error {
	return c.String(http.StatusOK, "sup")
}

func GetLoggedData(c echo.Context) error {
	pass := c.QueryParam("pass")
	if os.Getenv("QUERY_GENERAL_PASS") == pass {
		return c.File("./data/logs.json")

	} else {
		return c.JSON(echo.ErrUnauthorized.Code, &utils.ErrorMessage{
			Code:    echo.ErrUnauthorized.Code,
			Message: "Incorrect Credentials",
		})
	}
}
