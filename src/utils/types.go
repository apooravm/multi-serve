package utils

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ErrorMessage struct {
	Message     string `json:"message"`
	Description string `json:"description"`
}
