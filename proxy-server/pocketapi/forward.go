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

package pocketapi

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// This file contains helper functions for forwarding received API calls to a real Pocket instance (or a server with a compatible API).

func SendPocketRequest(host string, w http.ResponseWriter, r *http.Request, body string) {
	var bodyReader io.Reader = strings.NewReader(body)
	log.Printf("Sending request to https://%s%s", host, r.URL.Path)
	req, err := http.NewRequest(r.Method, fmt.Sprintf("https://%s%s", host, r.URL.Path), bodyReader)
	if err != nil {
		log.Printf("Unable to init request: %v", err)
		http.Error(w, "Error making request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", r.UserAgent())
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Unable to send request: %v", err)
		http.Error(w, "Error making request", http.StatusInternalServerError)
		return
	}

	log.Printf("Got response:")
	res.Write(log.Writer())
}
