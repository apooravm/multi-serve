package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ErrorMessage) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

type SuccessMessage struct {
	Message string `json:"message"`
}

type Log struct {
	ContentLength string
	Error         error
	Host          string
	Latency       time.Duration
	RemoteIP      string
	ResponseSize  int64
	Time          time.Time
	Status        int
	URI           string
	Protocol      string
}

type S3_File struct {
	LocalFilePath   string
	BucketObjectKey string
}

func S3_DownloadFiles() error {
	for _, file := range S3_Files {
		fileData, err := DownloadFile(BUCKET_NAME,
			file.BucketObjectKey,
			BUCKET_REGION)

		if err != nil {
			return &ServerError{
				Err:    err,
				Code:   echo.ErrInternalServerError.Code,
				Simple: "Error downloading file From bucket. Err: " + err.Error(),
			}
		}

		// If file exists, update it. Else create the dirs to the file
		if FileExists(file.LocalFilePath) {
			localFile, err := os.OpenFile(file.LocalFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return &ServerError{
					Err:    err,
					Code:   echo.ErrInternalServerError.Code,
					Simple: "Error opening file for writing. path: " + file.LocalFilePath,
				}

			}

			if _, err := localFile.Write(fileData); err != nil {
				return &ServerError{
					Err:    err,
					Code:   echo.ErrInternalServerError.Code,
					Simple: "Error writing to local file. Err: " + err.Error(),
				}
			}

			localFile.Close()

		} else {
			// Create the dirs leading to the file
			path_split := strings.Split(file.LocalFilePath, "/")
			mkdirPath := strings.Join(path_split[:len(path_split)-1], "/")
			err = os.MkdirAll(mkdirPath, os.ModePerm)
			if err != nil {
				panic(err)
			}

			// Writing the file to dir
			if err := os.WriteFile(file.LocalFilePath, fileData, 0644); err != nil {
				return &ServerError{
					Err:    err,
					Code:   echo.ErrInternalServerError.Code,
					Simple: "Error writing file data to local storage. path: " + file.LocalFilePath,
				}
			}
		}
	}

	return nil
}

func InitVars() {
	S3_Files = []S3_File{
		{
			LocalFilePath:   LOCAL_RESUME_PATH,
			BucketObjectKey: OBJ_RESUME_KEY,
		},
	}
}

var (
	S3_Files = []S3_File{}
)
