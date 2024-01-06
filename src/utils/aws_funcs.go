package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Download the resource file from the given s3 bucket
func DownloadFile(bucketName string, objPath string, region string) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	output, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objPath),
	})

	if err != nil {
		return nil, err
	}

	defer output.Body.Close()

	body, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func DownloadAllObjKeys(bucketName string, prefix string, region string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		return nil, err
	}

	var objKeyArr []string

	for _, item := range output.Contents {
		objKeyArr = append(objKeyArr, *item.Key)
	}

	return objKeyArr, nil
}

func DownloadAndWriteNoteData() error {
	objKeyArr, err := DownloadAllObjKeys(BUCKET_NAME, NOTES_DATA_FOLDER, BUCKET_REGION)
	if err != nil {
		return &ErrorMessage{
			Code:    500,
			Message: "Failed to Fetch Objects Keys. " + err.Error(),
		}
	}

	fmt.Println(objKeyArr)

	for _, objKey := range objKeyArr {
		file, err := DownloadFile(BUCKET_NAME, objKey, BUCKET_REGION)
		if err != nil {
			return &ErrorMessage{
				Code:    500,
				Message: "Error downloading the objects. " + err.Error(),
			}
		}
		localStoragePath := strings.ReplaceAll(objKey, "public/notes/", "")
		path_split := strings.Split(localStoragePath, "/")
		localStoragePath = "./data/notes/" + localStoragePath
		// path_split[path to file.txt]
		// [path to]
		// Join them, path/to
		createDirPath := "./data/notes/" + strings.Join(path_split[:len(path_split)-1], "/")

		// Create the required path
		err = os.MkdirAll(createDirPath, os.ModePerm)
		if err != nil {
			panic(err)
		}

		// So if the path is not a file, it throws an error
		// S3 returns dirs and files
		// dirs end with '/', thus "./data/notes/"
		// if end not / then create and write to the file
		// OR
		// relPath := c.Param("*")
		// basePath := "./data/notes"
		// absPath := filepath.Join(basePath, relPath)

		// // Check if file exists
		// _, err := os.Stat(absPath)
		// if err != nil {
		// 	return c.JSON(echo.ErrNotFound.Code, &utils.ErrorMessage{
		// 		Code:    echo.ErrNotFound.Code,
		// 		Message: "Requested file doesnt exist",
		// 	})
		// }
		if string(localStoragePath[len(localStoragePath)-1]) != "/" {
			if err := os.WriteFile(localStoragePath, file, 0644); err != nil {
				return &ErrorMessage{
					Code:    500,
					Message: "Error writing data to file. " + err.Error(),
				}
			}
		}
	}

	return nil
}
