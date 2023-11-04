package routes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func NotesGroup(group *echo.Group) {
	group.GET("/update", triggerPostDownload)
	group.GET("/info", GetRootInfo)
	group.GET("/:slug", GetMarkdownFromSlug)
	group.GET("/media/:file", GetNotesMedia)

}

func GetMarkdownFromSlug(c echo.Context) error {
	filename := c.Param("slug")
	if err := c.File("./data/notes/markdowns/" + filename); err != nil {
		return c.JSON(echo.ErrNotFound.Code, &utils.ErrorMessage{
			Code:    echo.ErrNotFound.Code,
			Message: "Requested file doesnt exist",
		})
	}
	return nil
}

func GetNotesMedia(c echo.Context) error {
	filename := c.Param("file")
	if err := c.File("./data/notes/media/" + filename); err != nil {
		return c.JSON(echo.ErrNotFound.Code, &utils.ErrorMessage{
			Code:    echo.ErrNotFound.Code,
			Message: "Requested file doesnt exist",
		})
	}
	return nil
}

func getAllFilePaths(postPath string, trimPrefix string) (*[]string, *utils.ErrorMessage) {
	var paths []string
	err := filepath.WalkDir(postPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return &utils.ErrorMessage{
				Code:    echo.ErrInternalServerError.Code,
				Message: "Error in post files",
			}
		}

		if !d.IsDir() {
			paths = append(paths, strings.TrimPrefix(filepath.ToSlash(path), trimPrefix))
		}

		return nil
	})

	if err != nil {
		return nil, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Error in accessing the local directories",
		}
	}

	return &paths, nil
}

// Does x thing
// very good at doing x thing
func findPostPath(rootPath string, postID string) (string, *utils.ErrorMessage) {
	var postPath string

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return &utils.ErrorMessage{
				Code:    echo.ErrInternalServerError.Code,
				Message: "Error in accessing the local directories",
			}
		}

		if d.IsDir() && path != rootPath {
			dir := filepath.Base(path)
			if strings.Contains(dir, postID) {
				postPath = path
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil {
		return "", &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Error in accessing the local directories",
		}
	}

	if postPath != "" {
		return postPath, nil
	}

	return "", &utils.ErrorMessage{
		Code:    echo.ErrNotFound.Code,
		Message: "Post with ID " + postID + " doesnt exist",
	}
}

func GetRootInfo(c echo.Context) error {
	infoJsonPath := utils.LOCAL_INFO_PATH
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

func triggerPostDownload(c echo.Context) error {
	qPass := c.QueryParam("pass")
	if utils.QUERY_TRIGGER_PASS == qPass {
		if err := utils.DownloadAndWriteNoteData(); err != nil {
			return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
				Code:    echo.ErrInternalServerError.Code,
				Message: "Server Error, Error fetching data from S3 Bucket: " + err.Error(),
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
