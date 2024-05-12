package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

// /clipboard
// Storing temporary data in json. Limit the number of items. Allow deletion
func WebClipboardGroup(group *echo.Group) {
	group.GET("", FetchClipboardData)
	group.POST("", PostClipboardData)
}

type ClipboardNew struct {
	Text string `json:"text"`
}

func PostClipboardData(c echo.Context) error {
	var newClipboardText ClipboardNew
	if err := c.Bind(&newClipboardText); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error binding:"+err.Error()))
	}

	if len(newClipboardText.Text) > 2000 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Text too long lol"))
	}

	readData, err := ReadWebClipboardData()
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr(err.Error()))
	}

	*readData = append(*readData, newClipboardText.Text)
	if err := WriteWebClipboardData(readData); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr(err.Error()))
	}

	return c.String(http.StatusOK, "")
}

func FetchClipboardData(c echo.Context) error {
	currentData, err := ReadWebClipboardData()
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr(err.Error()))
	}

	return c.JSON(http.StatusOK, currentData)
}

func WriteWebClipboardData(updatedData *[]string) error {
	newArr, err := json.MarshalIndent(updatedData, "", "    ")
	if err != nil {
		return fmt.Errorf("error encoding data. %s", err.Error())
	}

	if err := os.WriteFile(utils.CLIPBOARD_PATH, newArr, 0644); err != nil {
		return fmt.Errorf("error writing to file. %s", err.Error())
	}

	return nil
}

func ReadWebClipboardData() (*[]string, error) {
	var data []string

	file, err := os.Open(utils.CLIPBOARD_PATH)
	if err != nil {
		return nil, fmt.Errorf("error reading clipboard file. %s", err.Error())
	}

	defer file.Close()
	
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("error parsing clipboard file. %s", err.Error())
	}

	return &data, nil
}

// If file doesnt exist, it is created
// These are flags that control the behavior of the file opening operation
// os.O_WRONLY: This flag specifies that the file should be opened in write-only mode.
// os.O_APPEND: This flag specifies that data should be appended to the file when writing, rather than overwriting existing content.
// os.O_CREATE: This flag specifies that the file should be created if it doesn't exist.
// These flags are combined using the bitwise OR operator (|) to form a single integer value representing the desired options.
// 0644: This is the file mode or permission bits used when creating a new file (if os.O_CREATE flag is set). In Unix-like systems, file permissions are represented using octal notation. 0644 indicates that the file should have read and write permissions for the owner, and read-only permissions for others.

