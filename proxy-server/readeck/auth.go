// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
