package routes

import (
	"fmt"
	"mime"
	"path/filepath"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func S3FileFetchGroup(group *echo.Group) {
	group.GET("", GetFileFromObjectKey)
	// group.GET("/video", streamVideoHandler)
}

// Requires password and object key as query params
// Object key is checked againsts the ones fetched from bucket.
func GetFileFromObjectKey(c echo.Context) error {
	pass := c.QueryParam("pass")

	if utils.QUERY_GENERAL_PASS != pass {
		return c.JSON(echo.ErrUnauthorized.Code, &utils.ErrorMessage{
			Code:    echo.ErrUnauthorized.Code,
			Message: "Incorrect Credentials",
		})
	}

	object_key := c.QueryParam("object_key")

	object_keys, err := utils.GetObjectKeys()
	if err != nil {
		utils.LogData("Could not fetch object key list.", err.Error())
		return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Could not verify object key.",
		})
	}

	objectInfo := KeyExistsInList(object_key, object_keys)
	if objectInfo == nil {
		return c.JSON(echo.ErrBadRequest.Code, &utils.ErrorMessage{
			Code:    echo.ErrBadRequest.Code,
			Message: "Object Key not found in bucket.",
		})
	}

	if objectInfo.Size >= 50_000_000 {
		message := "File too big to download 😾"
		return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: fmt.Sprintf("%s [%.2fMB]", message, float64(objectInfo.Size)/float64(1000_000)),
		})
	}

	fileData, err := utils.DownloadFile(utils.BUCKET_NAME, object_key, utils.BUCKET_REGION)
	if err != nil {
		utils.LogData("Could not download file with object key", object_key, err.Error())
		return c.JSON(echo.ErrInternalServerError.Code, &utils.ErrorMessage{
			Code:    echo.ErrInternalServerError.Code,
			Message: "Could not download file.",
		})
	}

	// Get the appropriate content type depending on the type of file.
	// Learn more about this mime package.
	ext := filepath.Ext(object_key)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Response().Header().Set(echo.HeaderContentType, contentType)
	return c.Blob(200, contentType, fileData)
}

// Check whether the object key exists in the fetched object key list
func KeyExistsInList(objectKeyToCheck string, fetchedObjectKeys *[]utils.FileInfo) *utils.FileInfo {
	for _, key := range *fetchedObjectKeys {
		if key.ObjectKey == objectKeyToCheck {
			return &key
		}
	}

	return nil
}
