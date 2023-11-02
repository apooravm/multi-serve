package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

func AppendLogToFile(data *Log, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		fmt.Println("Failed to open log file")
		return err
	}

	defer file.Close()

	var logs []Log

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&logs); err != nil {
		if err != io.EOF {
			fmt.Println(err)
			return nil
		}
	}

	logs = append(logs, *data)

	updatedJSON, err := json.MarshalIndent(logs, "", "    ")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println(err)
		return nil
	}
	if err := file.Truncate(0); err != nil {
		fmt.Println(err)
		return nil
	}
	_, err = file.Write(updatedJSON)
	if err != nil {
		fmt.Println(err)
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
