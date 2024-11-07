package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// Check if a file exists at the given filepath.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Create the required directories
func InitDirs() {
	createDirPath := "./data/logs"
	err := os.MkdirAll(createDirPath, os.ModePerm)
	if err != nil {
		LogData("misc.go err_id:001 | Directories could not be initialized", err.Error())
	}
}

// Log text data to a give filepath. If the file doesnt exist, it is created.
// Variadic function that can take any number of data.
func LogDataToPath(logFilePath string, data ...string) {
	dataJoined := strings.Join(data, " ")
	// fmt.Println("DEVMODE_LOG:", dataJoined)
	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println("misc.go err_id:002 |", err.Error())
	}
	defer file.Close()

	currentTime := time.Now()
	timeString := currentTime.Format("2006-01-02 15:04:05")
	dataJoined = timeString + " " + dataJoined + "\n"

	_, err = file.WriteString(dataJoined)
	if err != nil {
		log.Println("misc.go err_id:003 |", err.Error())
		log.Println("Data that was being logged:", data)
	}
}

// Log data to the default file
func LogData(data ...string) {
	LogDataToPath(SERVER_LOG_PATH, data...)
}

// Difference between this and LogData is that this logs to a given file, while the other has a default filepath.
// Actually nvm idk
func AppendLogToFile(data *Log, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		LogData("misc.go err_id:004 | failed to open file", err.Error())
		return err
	}

	defer file.Close()

	var logs []Log

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&logs); err != nil {
		if err != io.EOF {
			LogData("misc.go err_id:006 |", err.Error())
			return nil
		}
	}

	logs = append(logs, *data)

	updatedJSON, err := json.MarshalIndent(logs, "", "    ")
	if err != nil {
		LogData("misc.go err_id:007 |", err.Error())
		return nil
	}

	if _, err := file.Seek(0, 0); err != nil {
		LogData("misc.go err_id:008 |", err.Error())
		return nil
	}
	if err := file.Truncate(0); err != nil {
		LogData("misc.go err_id:009 |", err.Error())
		return nil
	}
	_, err = file.Write(updatedJSON)
	if err != nil {
		LogData("misc.go err_id:010 |", err.Error())
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

func ClientErr(messageStr ...string) ErrorMessage {
	finalMessage := strings.Join(messageStr, " ")
	return ErrorMessage{
		Code:    echo.ErrBadRequest.Code,
		Message: finalMessage,
	}
}

// This file needs to be created beforehand unlike others because its json and has a structure.
func CreateWebClipboardFile() error {
	file, err := os.Create(CLIPBOARD_PATH_JSON)
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

func create_file(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}

func InitFiles() {
	if err := CreateWebClipboardFile(); err != nil {
		LogData("misc.go err_id:011 | error creating clipboard file", err.Error())
	}

	filepaths := []string{REQUEST_LOG_PATH, SERVER_LOG_PATH, DUMMY_WS_LOG_PATH, CHAT_DEBUG, CHAT_LOG, CLIPBOARD_PATH_TXT}

	for _, filepath := range filepaths {
		if err := create_file(filepath); err != nil {
			LogData("misc.go err_id:012 | error creating file at", filepath, err.Error())
		}
	}
}
