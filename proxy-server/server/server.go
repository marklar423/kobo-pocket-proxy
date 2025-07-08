package server

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"proxyserver/pocketapi"
	"proxyserver/readeck"
	"strings"
	"time"
)

var port = flag.Int("port", 8080, "HTTP port to listen on")
var verbose = flag.Bool("verbost", true, "If true, dumps all request fields to stdout")
var backendName = flag.String("backend", "readeck", "The name of the backend to forward API calls to")
var backendEndpoint = flag.String("backend_endpoint", "", "The backend API endpoint")
var backendBearerToken = flag.String("backend_bearer_token", "", "The backend API bearer token used for authentication")

type backendInit func() (Backend, error)

func initReadeck() (Backend, error) {
	if *backendEndpoint == "" {
		return nil, errors.New("need to specify --backend_endpoint for when using a Readeck backend")
	}
	if *backendBearerToken == "" {
		return nil, errors.New("need to specify --backend_bearer_token for when using a Readeck backend")
	}
	return readeck.NewReadeckConn(*backendEndpoint, *backendBearerToken), nil
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
}

func NewServer(backendName string) (*server, error) {
	backendInit, exists := allBackends[backendName]
	if !exists {
		return nil, fmt.Errorf("unknown backend \"%s\", available backends: %s", backendName, allBackendNames())
	}
	backend, err := backendInit()
	if err != nil {
		return nil, err
	}
	return &server{
		backend: backend,
	}, nil
}

func maybeVLog(r *http.Request, body string) {
	if *verbose {
		log.Printf("Got request at %s", r.URL)
		log.Println("--------------------------")
		for k, v := range r.Header {
			log.Printf("%s: %s", k, v)
		}
		log.Println(body)
		log.Println("--------------------------")
	}
}

func (s *server) getArticles(w http.ResponseWriter, r *http.Request) {
	maybeVLog(r, fmt.Sprintf("Received request %s", r.URL.String()))

	var body pocketapi.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse request body: %v", err.Error()), http.StatusBadRequest)
		return
	}

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
	maybeVLog(r, fmt.Sprintf("Received request %s", r.URL.String()))

	var body pocketapi.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse request body: %v", err.Error()), http.StatusBadRequest)
		return
	}

	var responseBody pocketapi.SendResponse
	responseBody.Status = 1
	responseBody.ActionResults = make([]bool, len(body.Actions))
	responseBody.ActionErrors = make([]*pocketapi.SendError, len(body.Actions))

	for i, action := range body.Actions {
		actionTime := time.Unix(int64(action.Time), 0)
		var actionErr error
		if action.Action == "add" {
			actionErr = s.backend.Add(action.URL, "", actionTime)
		} else if action.Action == "archive" {
			actionErr = s.backend.Archive(action.ItemID, actionTime)
		} else if action.Action == "readd" {
			actionErr = s.backend.Unarchive(action.ItemID, actionTime)
		} else if action.Action == "favorite" {
			actionErr = s.backend.Favorite(action.ItemID, actionTime)
		} else if action.Action == "unfavorite" {
			actionErr = s.backend.Unfavorite(action.ItemID, actionTime)
		} else if action.Action == "delete" {
			actionErr = s.backend.Delete(action.ItemID, actionTime)
		} else {
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
	maybeVLog(r, fmt.Sprintf("Received request %s", r.URL.String()))
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got unhandled request at %s", r.URL)
	http.NotFound(w, r)
}

func StartServing() {
	mux := http.NewServeMux()
	server, err := NewServer(*backendName)
	if err != nil {
		log.Printf("Error starting server: %v", err)
		return
	}

	mux.HandleFunc("/v3/get", server.getArticles)
	mux.HandleFunc("/v3/send", server.modifyArticles)
	mux.HandleFunc("/v3beta/text", server.articleText)
	mux.HandleFunc("/", catchAll)

	fmt.Printf("Listening on http://localhost:%d\n", *port)

	http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)
}
