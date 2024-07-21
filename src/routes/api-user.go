package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte("your_secret_key")

// api/user
func UserGroup(group *echo.Group) {
	group.POST("/register", registerNewUser)
	group.POST("/login", loginUser)
	group.POST("/getinfo", getUserInfo)
	group.POST("/verify", verifyToken, AuthJwtMiddleware)
}

// User flow
// Sign Up/Register at https://apooravm.vercel.app/register
// Login through browser or CLI apps and receive a jwt
// Store it and attach it to the Header as `Auth: "Bearer <token>"`
// Verified for each required route

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
		utils.LogData("Error Hashing" + err.Error())
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
		utils.LogData("Error updating db:", err.Error())
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error Updating db"))
	}

	req.Header.Set("apiKey", apiKey)
	req.Header.Set("Authorization", "Bearer"+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		utils.LogData("Error updating db 2", err.Error())
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error Updating db 2"))
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

// The claims are fields that allow control over the tokens validity or scope.
// For now using these to store basic user info
type JwtClaims struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Id       int    `json:"id"`
	jwt.StandardClaims
}

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func AuthJwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Auth")
		// Remove the Bearer prefix
		tokenString := strings.Split(authHeader, " ")[1]

		// Parse JWT token
		token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		// Invalid Token
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		// Valid Token
		if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
			c.Set("user", claims)
			return next(c)
		}

		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Token")
	}
}

func verifyToken(c echo.Context) error {
	user := c.Get("user").(*JwtClaims)
	return c.String(200, fmt.Sprintf("Hello %s", user.Username))
}

type UserAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func loginUser(c echo.Context) error {
	var newLogReq UserAuth
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Format"+err.Error()))
	}

	// Get user_id
	userProfiles, err := utils.GetUserFromEmail(newLogReq.Email)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr(err.Error()))
	}

	if len(userProfiles) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Credentials"))
	}

	// Auth Password
	userFromDB := userProfiles[0]
	if utils.ComparePasswords(userFromDB.Password, newLogReq.Password) != nil {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Password"))
	}

	// JWT Token gen
	token := jwt.New(jwt.SigningMethodHS256)
	claims := &JwtClaims{
		Email:    userFromDB.Email,
		Username: userFromDB.Username,
		Id:       userFromDB.Id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token.Claims = claims
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Something went wrong"+err.Error()))
	}

	return c.JSON(200, map[string]string{
		"token":    "Bearer " + t,
		"username": userFromDB.Username,
	})
}

type User2 struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Id       string `json:"id"`
}

func getUserInfo(c echo.Context) error {
	var newLogReq User2
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Format"+err.Error()))
	}

	return c.JSON(200, &newLogReq)
}
