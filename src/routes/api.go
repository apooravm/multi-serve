package routes

import (
	"net/http"

	"github.com/apooravm/multi-serve/src/routes/dummy_ws"
	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func ApiGroup(group *echo.Group) {
	group.GET("/resume", GetResume)
	group.GET("/resume/png", GetResumePNG)
	group.GET("/resume/html", GetResumeHTML)
	group.GET("/resume/pdf", GetResumePDF)
	group.GET("/ping", PingServer)
	group.GET("/logs", GetServerLogs)
	group.GET("/update", UpdateApiData)

	group.GET("/chat", Chat)
	group.GET("/chat/logs", GetChatLogs)
	group.GET("/chat/debug", GetChatDebug)

	group.GET("/ws/echo", dummy_ws.EchoDummyWS)

	NotesGroup(group.Group("/notes"))
	FileTransferGroup(group.Group("/filetransfer"))
	MiscGroup(group.Group("/misc"))
	JournalLoggerGroup(group.Group(("/journal")))
	UserGroup(group.Group("/user"))
}

// Refreshes the local server Data
func UpdateApiData(c echo.Context) error {
	qPass := c.QueryParam("pass")
	if qPass != utils.QUERY_TRIGGER_PASS {
		return c.JSON(echo.ErrBadRequest.Code, &utils.ErrorMessage{
			Code:    echo.ErrBadRequest.Code,
			Message: "Invalid Credentials",
		})
	}
	if err := utils.S3_DownloadFiles(); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Error Updating the Files. Error: " + err.Error(),
		})
	}

	return c.JSON(http.StatusAccepted, &utils.SuccessMessage{
		Message: "Local data updated and written successfully",
	})
}

func GetResumePDF(c echo.Context) error {
	return c.File(utils.LOCAL_RESUME_PATH)
}

// Initially download the file in ./data/files/resume.pdf
// If file exists at path then return the file
// Else download to the location
// Reduction in S3 fetching cost
func GetResume(c echo.Context) error {
	resumeFilePath := utils.LOCAL_RESUME_PATH

	// Downloading the file from Bucket
	// if !utils.FileExists(resumeFilePath) {
	// 	file, err := utils.DownloadFile(utils.BUCKET_NAME,
	// 		utils.OBJ_RESUME_KEY,
	// 		utils.BUCKET_REGION)

	// 	if err != nil {
	// 		errMsg := utils.ErrorMessage{
	// 			Code:    echo.ErrInternalServerError.Code,
	// 			Message: "Server Error, Error fetching data from S3 Bucket",
	// 		}
	// 		fmt.Println(errMsg)
	// 		fmt.Println("\nERR\n", err.Error())
	// 		return c.JSON(echo.ErrInternalServerError.Code, &errMsg)
	// 	}

	// 	path_split := strings.Split(resumeFilePath, "/")
	// 	mkdirPath := strings.Join(path_split[:len(path_split)-1], "/")
	// 	err = os.MkdirAll(mkdirPath, os.ModePerm)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	// Writing the file to dir
	// 	if err := os.WriteFile(resumeFilePath, file, 0644); err != nil {
	// 		return c.JSON(echo.ErrInternalServerError.Code, utils.ErrorMessage{
	// 			Code:    echo.ErrInternalServerError.Code,
	// 			Message: "Serve Error, Error writing to local file. " + err.Error(),
	// 		})
	// 	}
	// }
	return c.File(resumeFilePath)
}

func GetResumePNG(c echo.Context) error {
	resumeFilePath := utils.LOCAL_RESUME_PNG_PATH
	return c.File(resumeFilePath)
}

func GetResumeHTML(c echo.Context) error {
	resumeFilePath := utils.LOCAL_RESUME_HTML_PATH
	return c.File(resumeFilePath)
}

func GetServerLogs(c echo.Context) error {
	pass := c.QueryParam("pass")

	if utils.QUERY_GENERAL_PASS == pass {
		return c.File(utils.SERVER_LOG_PATH)

	} else {
		return c.JSON(echo.ErrUnauthorized.Code, &utils.ErrorMessage{
			Code:    echo.ErrUnauthorized.Code,
			Message: "Incorrect Credentials",
		})
	}
}

func GetChatDebug(c echo.Context) error {
	pass := c.QueryParam("pass")
	if utils.QUERY_GENERAL_PASS == pass {
		return c.File(utils.CHAT_DEBUG)
	} else {
		return c.JSON(echo.ErrUnauthorized.Code, &utils.ErrorMessage{
			Code:    echo.ErrUnauthorized.Code,
			Message: "Incorrect Credentials",
		})
	}
}

func GetChatLogs(c echo.Context) error {
	pass := c.QueryParam("pass")
	if utils.QUERY_GENERAL_PASS == pass {
		return c.File(utils.CHAT_LOG)
	} else {
		return c.JSON(echo.ErrUnauthorized.Code, &utils.ErrorMessage{
			Code:    echo.ErrUnauthorized.Code,
			Message: "Incorrect Credentials",
		})
	}
}

func PingServer(c echo.Context) error {
	return c.String(http.StatusOK, "sup")
}

func GetLoggedData(c echo.Context) error {
	pass := c.QueryParam("pass")
	if utils.QUERY_GENERAL_PASS == pass {
		return c.File("./data/logs.json")

	} else {
		return c.JSON(echo.ErrUnauthorized.Code, &utils.ErrorMessage{
			Code:    echo.ErrUnauthorized.Code,
			Message: "Incorrect Credentials",
		})
	}
}
