package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Create the required directories
func InitDirs() {
	createDirPath := "./data/logs"
	err := os.MkdirAll(createDirPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

// Logs Data into log file. If file doesnt exist, it is created.
func LogData(data string, logFilePath string) {
	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		moreErr := ServerError{
			Err:    err,
			Code:   SERVER_ERR,
			Simple: "Error opening the log file",
		}
		fmt.Println("misc.go ln:36 |", moreErr.Error())
	}
	defer file.Close()

	currentTime := time.Now()
	timeString := currentTime.Format("2006-01-02 15:04:05")
	data = timeString + " " + data + "\n"

	_, err = file.WriteString(data)
	if err != nil {
		moreErr := ServerError{
			Err:    err,
			Code:   SERVER_ERR,
			Simple: "Error logging",
		}
		fmt.Println("misc.go ln:51 |", moreErr.Error())
		fmt.Println("Data that was being logged:", data)
	}
}

func AppendLogToFile(data *Log, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		fmt.Println("misc.go ln:58 |", "Failed to open log file")
		return err
	}

	defer file.Close()

	var logs []Log

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&logs); err != nil {
		if err != io.EOF {
			fmt.Println("misc.go ln:69 |", err)
			return nil
		}
	}

	logs = append(logs, *data)

	updatedJSON, err := json.MarshalIndent(logs, "", "    ")
	if err != nil {
		fmt.Println("misc.go ln:78 |", err)
		return nil
	}

	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println("misc.go ln:83 |", err)
		return nil
	}
	if err := file.Truncate(0); err != nil {
		fmt.Println("misc.go ln:87 |", err)
		return nil
	}
	_, err = file.Write(updatedJSON)
	if err != nil {
		fmt.Println("misc.go ln:92 |", err)
		return nil
	}

	return nil
}

// type GlobalMapper struct {
// 	ctx context.Context
// 	key string
// }

// func newGlobalMapper() *GlobalMapper {
// 	return &GlobalMapper{
// 		key: "GMapper",
// 	}
// }

// func (gm *GlobalMapper) AddValue(key string, value interface{}) {
// 	// Retrieve dataMap
// 	dataMap, ok := gm.ctx.Value(gm.key).(map[string]interface{})
// 	if !ok {
// 		// dataMap doesnt exist
// 		dataMap = make(map[string]interface{})
// 	}
// 	dataMap[key] = value

// 	// Store the dataMap back
// 	gm.ctx = context.WithValue(gm.ctx, gm.key, dataMap)
// }

// func (gm *GlobalMapper) GetValue(key string) any {
// 	// Retrieve dataMap
// 	dataMap, ok := gm.ctx.Value(gm.key).(map[string]interface{})
// 	if !ok {
// 		// dataMap doesnt exist
// 		dataMap = make(map[string]interface{})
// 	}
// 	return dataMap[key]
// }

func InternalServerErr(messageStr string) ErrorMessage {
	return ErrorMessage{
		Code:    echo.ErrInternalServerError.Code,
		Message: messageStr,
	}
}

func ClientErr(messageStr string) ErrorMessage {
	return ErrorMessage{
		Code:    echo.ErrBadRequest.Code,
		Message: messageStr,
	}
}

func CreateWebClipboardFile() error {
	file, err := os.Create(CLIPBOARD_PATH)
	if err != nil {
		return fmt.Errorf("error creating webClipboard file. %s", err.Error())
	}

	clipboardInit := make([]string, 0)
	clipboardData, err := json.MarshalIndent(&clipboardInit, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling initial clipboard data. %s", err.Error())
	}

	_, err = file.Write(clipboardData)
	if err != nil {
		return fmt.Errorf("error writing to clipboard file. %s", err.Error())
	}

	return nil
}

func InitFiles() {
	if err := CreateWebClipboardFile(); err != nil {
		LogData(err.Error(), SERVER_LOG_PATH)
	}
}
