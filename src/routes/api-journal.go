package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

// /api/journal
func JournalLoggerGroup(group *echo.Group) {
	group.GET("/log", GetJournalLogs)
	group.POST("/", PostJournalLogEntry)
}

func GetJournalLogs(c echo.Context) error {
	limit := c.QueryParam("limit")

	var newLogReq UserLogReq
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Credential Format"+err.Error()))
	}

	// Get user_id
	userProfiles, err := getUserFromUsername(newLogReq.Username)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("DB Error. "+err.Error()))
	}

	if len(userProfiles) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Credentials"))
	}

	// Auth Password
	userFromDB := userProfiles[0]
	if userFromDB.Password != newLogReq.Password {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Password"))
	}

	url := utils.DB_URL + "userlog?select=created_at,log_message"
	apiKey := utils.DB_API_KEY

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Creating Request"))
	}

	req.Header.Set("apiKey", apiKey)
	req.Header.Set("Authorization", "Bearer"+apiKey)
	req.Header.Set("Content-Type", "application/json")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Limit"))
	}

	if len(limit) != 0 && limitInt > -1 {
		req.Header.Set("Range", "0-"+strconv.Itoa(limitInt-1))
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Sending Request. "+err.Error()))
	}

	defer res.Body.Close()

	var resLog []UserLogRes

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if err := json.NewDecoder(res.Body).Decode(&resLog); err != nil {
			return c.JSON(echo.ErrInternalServerError.Code,
				utils.InternalServerErr("Error Unmarshalling Request"+err.Error()))
		}

	} else {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Something went wrong"+err.Error()))
	}

	return c.JSON(200, &resLog)
}

type UserLogRes struct {
	Created_at string `json:"created_at"`
	Log        string `json:"log_message"`
}

type UserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLogReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Log      string `json:"log"`
}

type UserLogDb struct {
	User_id int    `json:"user_id"`
	Log     string `json:"log_message"`
}

func PostJournalLogEntry(c echo.Context) error {
	var newLogReq UserLogReq
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Credential Format"+err.Error()))
	}

	// Get user_id
	userProfiles, err := getUserFromUsername(newLogReq.Username)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Updating db. "+err.Error()))
	}

	if len(userProfiles) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Credentials"))
	}

	// Auth Password
	userFromDB := userProfiles[0]
	if utils.ComparePasswords(userFromDB.Password, newLogReq.Password) != nil {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Password"))
	}

	newLog := UserLogDb{
		User_id: userFromDB.Id,
		Log:     newLogReq.Log,
	}

	url := utils.DB_URL + "userlog"
	apiKey := utils.DB_API_KEY

	logBytes, err := json.Marshal(&newLog)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Marshalling"))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(logBytes))
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Creating Request"))
	}

	req.Header.Set("apiKey", apiKey)
	req.Header.Set("Authorization", "Bearer"+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Sending Request"))
	}

	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return c.JSON(res.StatusCode, utils.SuccessMessage{
			Message: "Log Created Successfully",
		})

	} else {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Something went wrong. "+res.Status))
	}
}

func getUserFromUsername(username string) ([]utils.UserProfile, error) {
	url := utils.DB_URL + "userprofile?username=eq." + username + "&select=id,username,password,email"
	apiKey := utils.DB_API_KEY

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []utils.UserProfile{}, &utils.ServerError{
			Err:    err,
			Code:   500,
			Simple: "Error Creating Request",
		}
	}

	req.Header.Set("apiKey", apiKey)
	req.Header.Set("Authorization", "Bearer"+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	// Create an http client and send req
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return []utils.UserProfile{}, &utils.ServerError{
			Err:    err,
			Code:   500,
			Simple: "Error Sending Request",
		}
	}

	defer res.Body.Close()

	var userProfiles []utils.UserProfile

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if err := json.NewDecoder(res.Body).Decode(&userProfiles); err != nil {
			return []utils.UserProfile{}, &utils.ServerError{
				Err:    err,
				Code:   500,
				Simple: "Error Sending Request",
			}
		}

	} else {
		return []utils.UserProfile{}, &utils.ServerError{
			Err:    err,
			Code:   500,
			Simple: "Something went wrong. " + res.Status,
		}
	}

	return userProfiles, nil
}
