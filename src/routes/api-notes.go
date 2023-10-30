package routes

import (
	"encoding/json"
	"os"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func NotesGroup(group *echo.Group) {
	group.GET("/update", triggerNotesDownload)
	group.GET("/info", GetRootInfo)

}

func GetRootInfo(c echo.Context) error {
	infoJsonPath := os.Getenv("LOCAL_INFO_PATH")
	data, err := os.ReadFile(infoJsonPath)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Error reading the local info file",
		})
	}

	// Parsing the Jsond data
	var infoJsonData interface{}
	if err := json.Unmarshal(data, &infoJsonData); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Error parsing the local info file",
		})
	}

	return c.JSON(200, infoJsonData)
}

func triggerNotesDownload(c echo.Context) error {
	qPass := c.QueryParam("pass")
	if os.Getenv("QUERY_TRIGGER_PASS") == qPass {
		if err := utils.DownloadAndWriteNoteData(); err != nil {
			return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
				Code:    echo.ErrInternalServerError.Code,
				Message: "Server Error, Error fetching data from S3 Bucket",
			})
		}
		return c.JSON(200, &utils.SuccessMessage{
			Message: "Data was written successfully!",
		})

	} else {
		return c.JSON(echo.ErrBadRequest.Code, &utils.ErrorMessage{
			Code:    401,
			Message: "Incorrect Credentials",
		})
	}
}
