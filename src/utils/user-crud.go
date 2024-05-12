package utils

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		LogData("Error hashing password", SERVER_LOG_PATH)
		return "", err
	}

	return string(hashedPass), nil
}

func ComparePasswords(hashedPassword string, enteredPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(enteredPassword))
}

// Returns user array given the email.
// If the provided email not in the db, returns an empty array.
// This allows to check whether the user exists or no.
// Remember to account for offline db server.
func GetUserFromEmail(email string) ([]UserProfile, error) {
	url := DB_URL + "userprofile?email=eq." + email + "&select=id,username,password,email"
	apiKey := DB_API_KEY

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []UserProfile{}, &ServerError{
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
	// DB server is offline. Thus need to send special message.
	// Or maybe some issue with my own connection.
	// Not checking that for now.
	if err != nil {
		return []UserProfile{}, &ServerError{
			Err:    err,
			Code:   500,
			Simple: "DB Server is offline. Try again later.",
		}
	}

	defer res.Body.Close()

	var userProfiles []UserProfile

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if err := json.NewDecoder(res.Body).Decode(&userProfiles); err != nil {
			return []UserProfile{}, &ServerError{
				Err:    err,
				Code:   500,
				Simple: "Error Sending Request",
			}
		}

	} else {
		return []UserProfile{}, &ServerError{
			Err:    err,
			Code:   500,
			Simple: "Something went wrong. " + res.Status,
		}
	}

	return userProfiles, nil
}
