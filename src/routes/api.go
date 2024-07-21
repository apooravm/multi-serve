package routes

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/apooravm/multi-serve/src/routes/dummy_ws"
	untitledgame "github.com/apooravm/multi-serve/src/routes/untitled-game"
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

	group.GET("/game", untitledgame.UntitledGameSocket)
	group.GET("/video", NormalVideoStream)
	group.GET("/chunkedvideo", ChunkedVideoStream)

	S3FileFetchGroup(group.Group("/files"))
	S3FileFetchGroup(group.Group("/files"))
	NotesGroup(group.Group("/notes"))
	FileTransferGroup(group.Group("/filetransfer"))
	MiscGroup(group.Group("/misc"))
	JournalLoggerGroup(group.Group(("/journal")))
	UserGroup(group.Group("/user"))
	WebClipboardGroup(group.Group("/clipboard"))
}

func NormalVideoStream(c echo.Context) error {
	return c.File("./local/sample_vid.mp4")
}

func ChunkedVideoStream(c echo.Context) error {
	filePath := "./local/sample_vid.mp4"
	file, err := os.Open(filePath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to open file")
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get file info")
	}

	fileSize := fileInfo.Size()
	rangeHeader := c.Request().Header.Get("Range")
	if rangeHeader == "" {
		return c.File(filePath)
	}

	parts := strings.Split(rangeHeader, "=")
	if len(parts) != 2 || parts[0] != "bytes" {
		return c.String(http.StatusBadRequest, "Invalid range header")
	}

	rangeParts := strings.Split(parts[1], "-")
	start, err := strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid range start")
	}

	var end int64
	if len(rangeParts) > 1 && rangeParts[1] != "" {
		end, err = strconv.ParseInt(rangeParts[1], 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid range end")
		}
	} else {
		end = fileSize - 1
	}

	if start > end || end >= fileSize {
		return c.String(http.StatusRequestedRangeNotSatisfiable, "Requested range not satisfiable")
	}

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to seek file")
	}

	buf := make([]byte, 4096) // Adjust the buffer size as needed
	c.Response().Header().Set(echo.HeaderContentType, "video/mp4")
	c.Response().Header().Set("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(fileSize, 10))
	c.Response().Header().Set("Accept-Ranges", "bytes")
	c.Response().WriteHeader(http.StatusPartialContent)

	for {
		toRead := int64(len(buf))
		if end-start+1 < toRead {
			toRead = end - start + 1
		}

		n, err := file.Read(buf[:toRead])
		if err != nil && err != io.EOF {
			return c.String(http.StatusInternalServerError, "Error reading file")
		}
		if n == 0 {
			break
		}

		if _, err := c.Response().Write(buf[:n]); err != nil {
			return c.String(http.StatusInternalServerError, "Error writing response")
		}

		c.Response().Flush()
		start += int64(n)
	}
	return nil
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
