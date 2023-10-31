package utils

import (
	"fmt"
	"time"
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
