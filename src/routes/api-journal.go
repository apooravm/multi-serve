package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/apooravm/multi-serve/src/utils"
	"github.com/labstack/echo/v4"
)

// /api/journal
func JournalLoggerGroup(group *echo.Group) {
	group.GET("/log", GetJournalLogs)
	group.POST("/", PostJournalLogEntry)
	group.GET("/", GetUserLogs)
	group.PUT("/", UpdateJournalLog)
	group.DELETE("/", DeleteJournalLog)
}

func GetJournalLogs(c echo.Context) error {
	limit := c.QueryParam("limit")

	var newLogReq ClientUserLogReq
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Credential Format"+err.Error()))
	}

	// Get user_id
	userProfiles, err := utils.GetUserFromEmail(newLogReq.Username)
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
			utils.InternalServerErr("Something went wrong"))
	}

	return c.JSON(200, &resLog)
}

type UserAuthField struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLogRes struct {
	Created_at string   `json:"created_at"`
	Log        string   `json:"log_message"`
	Title      string   `json:"title"`
	Tags       []string `json:"tags"`
	Log_Id     int      `json:"log_id"`
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LogInfo struct {
	Log     string   `json:"log"`
	Title   string   `json:"title"`
	Tags    []string `json:"tags"`
	User_id int      `json:"user_id"`
	Log_Id  int      `json:"log_id"`
}

type ClientUserLogReq struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Log      string   `json:"log"`
	Tags     []string `json:"tags"`
	Title    string   `json:"title"`
}

type NewUserLogCreate struct {
	User_id int      `json:"user_id"`
	Log     string   `json:"log_message"`
	Tags    []string `json:"tags"`
	Title   string   `json:"title"`
}

type UpdateLogReq struct {
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Log      string   `json:"log"`
	Tags     []string `json:"tags"`
	Title    string   `json:"title"`
	Log_Id   int      `json:"log_id"`
}

type UpdateLogReqDB struct {
	Log   string   `json:"log_message"`
	Tags  []string `json:"tags"`
	Title string   `json:"title"`
}

func PostJournalLogEntry(c echo.Context) error {
	var newLogReq ClientUserLogReq
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Format"+err.Error()))
	}

	// Get user_id
	userProfiles, err := utils.GetUserFromEmail(newLogReq.Username)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error reading DB. "+err.Error()))
	}

	if len(userProfiles) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Credentials"))
	}

	// Auth Password
	userFromDB := userProfiles[0]
	if utils.ComparePasswords(userFromDB.Password, newLogReq.Password) != nil {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Password"))
	}

	if len(newLogReq.Log) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Log Contents cannot be empty"))
	}

	if len(newLogReq.Tags) == 0 {
		newLogReq.Tags = []string{""}
	}

	if len(newLogReq.Title) == 0 {
		newLogReq.Title = ""
	}

	newLog := NewUserLogCreate{
		User_id: userFromDB.Id,
		Log:     newLogReq.Log,
		Tags:    newLogReq.Tags,
		Title:   newLogReq.Title,
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

func UpdateJournalLog(c echo.Context) error {
	var newLogReq UpdateLogReq
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Format"+err.Error()))
	}

	// Get user_id
	userProfiles, err := utils.GetUserFromEmail(newLogReq.Username)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error reading DB. "+err.Error()))
	}

	if len(userProfiles) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Credentials"))
	}

	// Auth Password
	userFromDB := userProfiles[0]
	if utils.ComparePasswords(userFromDB.Password, newLogReq.Password) != nil {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Password"))
	}

	// Update the log where log_id
	// url := utils.DB_URL + "userprofile?username=eq." + username + "&select=id,username,password,email"
	url := utils.DB_URL + "userlog?" + "log_id=eq." + strconv.Itoa(newLogReq.Log_Id) + "&user_id=eq." + strconv.Itoa(userFromDB.Id)
	apiKey := utils.DB_API_KEY

	var DBReqBody UpdateLogReqDB
	DBReqBody.Log = newLogReq.Log
	DBReqBody.Tags = newLogReq.Tags
	DBReqBody.Title = newLogReq.Title

	logBytes, err := json.Marshal(&DBReqBody)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Marshalling"))
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(logBytes))
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

func GetUserLogs(c echo.Context) error {
	var newLogReq UserAuthField
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Format"+err.Error()))
	}

	// Get user_id
	userProfiles, err := utils.GetUserFromEmail(newLogReq.Email)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error reading DB. "+err.Error()))
	}

	if len(userProfiles) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Credentials"))
	}

	// Auth Password
	userFromDB := userProfiles[0]
	if utils.ComparePasswords(userFromDB.Password, newLogReq.Password) != nil {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Password"))
	}

	url := utils.DB_URL + "userlog?user_id=eq." + strconv.Itoa(userFromDB.Id) + "&select=created_at,log_id,log_message,title,tags"
	apiKey := utils.DB_API_KEY

	req, err := http.NewRequest("GET", url, nil)
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

	var userLogs []UserLogRes

	if err := json.NewDecoder(res.Body).Decode(&userLogs); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Error decoding DB response. "+res.Status))
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return c.JSON(res.StatusCode, &userLogs)

	} else {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Something went wrong server side. "+res.Status))
	}
}

type DeleteLogReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Log_Id   int    `json:"log_id"`
}

func DeleteJournalLog(c echo.Context) error {
	var newLogReq DeleteLogReq
	if err := c.Bind(&newLogReq); err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Invalid Format"+err.Error()))
	}

	// Get user_id
	userProfiles, err := utils.GetUserFromEmail(newLogReq.Email)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error reading DB. "+err.Error()))
	}

	if len(userProfiles) == 0 {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Credentials"))
	}

	// Auth Password
	userFromDB := userProfiles[0]
	if utils.ComparePasswords(userFromDB.Password, newLogReq.Password) != nil {
		return c.JSON(echo.ErrBadRequest.Code, utils.ClientErr("Invalid Password"))
	}

	url := utils.DB_URL + "userlog?" + "log_id=eq." + strconv.Itoa(newLogReq.Log_Id) + "&user_id=eq." + strconv.Itoa(userFromDB.Id)
	apiKey := utils.DB_API_KEY

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Creating Request"))
	}

	req.Header.Set("apiKey", apiKey)
	req.Header.Set("Authorization", "Bearer"+apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return c.JSON(echo.ErrInternalServerError.Code,
			utils.InternalServerErr("Error Sending Request"))
	}

	defer res.Body.Close()
	fmt.Println(res.Status)

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return c.JSON(res.StatusCode, utils.SuccessMessage{
			Message: "Log Deleted Successfully",
		})

	} else {
		return c.JSON(echo.ErrInternalServerError.Code, utils.InternalServerErr("Something went wrong. "+res.Status))
	}

}
