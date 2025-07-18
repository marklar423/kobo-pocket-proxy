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

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"proxyserver/pocketapi"
	"proxyserver/readeck"
	"strings"
	"time"
)

type Options interface {
	Port() int
	Verbose() bool
	BackendName() string
	BackendEndpoint() string
	BackendBearerToken() string
}

type backendInit func(Options) (Backend, error)

func initReadeck(options Options) (Backend, error) {
	if options.BackendEndpoint() == "" {
		return nil, errors.New("need to specify --backend_endpoint when using a Readeck backend")
	}
	if options.BackendBearerToken() == "" {
		return nil, errors.New("need to specify --backend_bearer_token when using a Readeck backend")
	}
	return readeck.NewReadeckConn(options.BackendEndpoint(), options.BackendBearerToken()), nil
}

var allBackends = map[string]backendInit{
	"readeck": initReadeck,
}

func allBackendNames() string {
	names := make([]string, 0, len(allBackends))
	for k := range allBackends {
		names = append(names, k)
	}
	return strings.Join(names, ", ")
}

type server struct {
	backend Backend
	options Options
}

func NewServer(options Options) (*server, error) {
	backendInit, exists := allBackends[options.BackendName()]
	if !exists {
		return nil, fmt.Errorf("unknown backend \"%s\", available backends: %s", options.BackendName(), allBackendNames())
	}
	backend, err := backendInit(options)
	if err != nil {
		return nil, err
	}
	return &server{
		backend: backend,
		options: options,
	}, nil
}

func (s *server) log(r *http.Request) {
	log.Printf("%s %s received from %s", r.Method, r.URL, r.RemoteAddr)
	if s.options.Verbose() {
		for k, v := range r.Header {
			log.Printf("%s: %s", k, v)
		}
	}
}

func (s *server) logStruct(r *http.Request, data interface{}) {
	if s.options.Verbose() {
		log.Printf("%s: %+v", r.URL.Path, data)
	}
}

func (s *server) getArticles(w http.ResponseWriter, r *http.Request) {
	s.log(r)

	var body pocketapi.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse request body: %v", err), http.StatusBadRequest)
		return
	}
	s.logStruct(r, body)

	responseBody, err := s.backend.Get(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to forward request: %v", err), http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(w).Encode(&responseBody); err != nil {
		http.Error(w, fmt.Sprintf("Unable to serialize response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *server) modifyArticles(w http.ResponseWriter, r *http.Request) {
	s.log(r)

	var body pocketapi.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse request body: %v", err), http.StatusBadRequest)
		return
	}
	s.logStruct(r, body)

	var responseBody pocketapi.SendResponse
	responseBody.Status = 1
	responseBody.ActionResults = make([]bool, len(body.Actions))
	responseBody.ActionErrors = make([]*pocketapi.SendError, len(body.Actions))

	for i, action := range body.Actions {
		actionTime := time.Unix(int64(action.Time), 0)
		var actionErr error
		switch action.Action {
		case "add":
			actionErr = s.backend.Add(action.URL, "", actionTime)
		case "archive":
			actionErr = s.backend.Archive(action.ItemID, actionTime)
		case "readd":
			actionErr = s.backend.Unarchive(action.ItemID, actionTime)
		case "favorite":
			actionErr = s.backend.Favorite(action.ItemID, actionTime)
		case "unfavorite":
			actionErr = s.backend.Unfavorite(action.ItemID, actionTime)
		case "delete":
			actionErr = s.backend.Delete(action.ItemID, actionTime)
		default:
			// Do nothing, fail open.
			actionErr = nil // For emphasis.
		}

		responseBody.ActionResults[i] = (actionErr == nil)
		if actionErr != nil {
			responseBody.Status = 0
			responseBody.ActionErrors[i] = &pocketapi.SendError{
				Message: fmt.Sprintf("Unable to forward %s request: %v", action.Action, actionErr),
			}
		}
	}
	if err := json.NewEncoder(w).Encode(&responseBody); err != nil {
		http.Error(w, fmt.Sprintf("Unable to serialize response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *server) articleText(w http.ResponseWriter, r *http.Request) {
	s.log(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse request body: %v", err), http.StatusBadRequest)
		return
	}

	url, exists := r.Form["url"]
	if !exists || len(url) == 0 {
		http.Error(w, "No URL specified in form data", http.StatusBadRequest)
		return
	}

	responseBody, err := s.backend.ArticleText(url[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to forward request: %v", err), http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(w).Encode(&responseBody); err != nil {
		http.Error(w, fmt.Sprintf("Unable to serialize response: %v", err), http.StatusInternalServerError)
		return
	}
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got unhandled request at %s", r.URL)
	http.NotFound(w, r)
}

func StartServing(options Options) {
	mux := http.NewServeMux()
	server, err := NewServer(options)
	if err != nil {
		log.Printf("Error starting server: %v", err)
		return
	}

	mux.HandleFunc("/v3/get", server.getArticles)
	mux.HandleFunc("/v3/send", server.modifyArticles)
	mux.HandleFunc("/v3beta/text", server.articleText)
	mux.HandleFunc("/", catchAll)

	fmt.Printf("Listening on http://localhost:%d\n", options.Port())

	err = http.ListenAndServe(fmt.Sprintf(":%d", options.Port()), mux)
	fmt.Printf("Server: %v", err)
}
