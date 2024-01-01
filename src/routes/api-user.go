package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

func UserGroup(group *echo.Group) {
	group.POST("/register", registerNewUser)
}

// No need to check against the user table
// If conflict, return it back
func registerNewUser(c echo.Context) error {
	// Reading and validating request body
	var newUser utils.UserRegister

	if err := c.Bind(&newUser); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error binding:"+err.Error()))
	}

	if len(newUser.Email) == 0 || len(newUser.Username) == 0 || len(newUser.Password) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Incomplete Credentials"))
	}

	// Marshalling newUser to string array
	jsonBytes, err := json.Marshal(newUser)
	if err != nil {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Error reading request contents:"+err.Error()))
	}

	url := os.Getenv("DB_URL") + "userprofile"
	apiKey := os.Getenv("DB_KEY")

	// Creating and sending request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error Updating db: "+err.Error()))
	}

	req.Header.Set("apiKey", apiKey)
	req.Header.Set("Authorization", "Bearer"+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error Updating db 2: "+err.Error()))
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 409:
		return c.JSON(echo.ErrConflict.Code, &utils.ErrorMessage{
			Code:    echo.ErrConflict.Code,
			Message: "User already exists. Pick a different username or email",
		})

	case 201:
		return c.JSON(201, &utils.SuccessMessage{Message: "Success! " + res.Status})

	default:
		return c.JSON(201, &utils.SuccessMessage{Message: "Success! " + res.Status})
	}
}

/*
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type UserProfile struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func getAllUsers() {
	url := "https://szvftfyqupyrjrqpvrsr.supabase.co/rest/v1/userprofile?username=eq.mrBruh&select=id,username,password,email"
	apiKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InN6dmZ0ZnlxdXB5cmpycXB2cnNyIiwicm9sZSI6ImFub24iLCJpYXQiOjE2OTk2NDM0NDMsImV4cCI6MjAxNTIxOTQ0M30.kqkmEq2DEkJu8ymc_qcHQDlUknRsqghZuyTolJ7IOzg"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request", err)
		return
	}

	req.Header.Set("apiKey", apiKey)
	req.Header.Set("Authorization", "Bearer"+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	// Create an http client and send req
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error Sending request", err)
		return
	}

	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {

		// Read and print res body
		var userProfiles []UserProfile

		if err := json.NewDecoder(res.Body).Decode(&userProfiles); err != nil {
			fmt.Println("Error Decoding to json", err)
			return
		}

		// If db returns empty arr
		if len(userProfiles) == 0 {
			fmt.Println("Invalid credentials")
			return
		}

		profile := userProfiles[0]
		fmt.Println(profile)

	} else {
		fmt.Println("Something went wrong", res.Status)
	}
}

func main() {
	getAllUsers()

	// url := "https://szvftfyqupyrjrqpvrsr.supabase.co/rest/v1/userprofile"
	// apiKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InN6dmZ0ZnlxdXB5cmpycXB2cnNyIiwicm9sZSI6ImFub24iLCJpYXQiOjE2OTk2NDM0NDMsImV4cCI6MjAxNTIxOTQ0M30.kqkmEq2DEkJu8ymc_qcHQDlUknRsqghZuyTolJ7IOzg"

	// // JSON payload
	// payload := []byte(`{"username": "AnotherOne", "email": "MrBigManCrones@mail.com", "password": "dontguesspls"}`)

	// req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	// if err != nil {
	// 	fmt.Println("Error creating request", err)
	// 	return
	// }

	// req.Header.Set("apiKey", apiKey)
	// req.Header.Set("Authorization", "Bearer"+apiKey)
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Prefer", "return=minimal")

	// // Create an http client and send req
	// client := &http.Client{}
	// res, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error Sending request", err)
	// 	return
	// }

	// defer res.Body.Close()

	// fmt.Println("Res Status", res.Status)

	// // Read and print res body
	// buf := new(bytes.Buffer)
	// buf.ReadFrom(res.Body)
	// fmt.Println("Res Body", buf.String())
}

*/
