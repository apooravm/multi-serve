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

	hashedPass, err := utils.HashPassword(newUser.Password)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error Hashing"+err.Error()))
	}

	newUser.Password = hashedPass

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
