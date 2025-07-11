package readeck

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func GetAuthToken(baseUrl string, appName, username, password string) (string, error) {
	url := fmt.Sprintf("%s/api/auth", baseUrl)
	payload := []byte(fmt.Sprintf(`{
		"application": "%s",
		"username": "%s",
		"password": "%s"
	}`, appName, username, password))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if err := checkResponseCode(res); err != nil {
		return "", err
	}

	var resBody struct {
		Token string
	}
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return "", err
	}
	if resBody.Token == "" {
		return "", errors.New("unexpected empty token in Readeck auth response")
	}

	return resBody.Token, nil
}
