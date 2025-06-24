package server

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"proxyserver/backend"
	"strings"
)

type backendInit func() (backend.Backend, error)

var port = flag.Int("port", 8080, "HTTP port to listen on")
var verbose = flag.Bool("verbost", true, "If true, dumps all request fields to stdout")
var backendName = flag.String("backend", "readeck", "The name of the backend to forward API calls to")
var backendEndpoint = flag.String("backend_endpoint", "", "The backend API endpoint")
var backendBearerToken = flag.String("backend_bearer_token", "", "The backend API bearer token used for authentication")

func initReadeck() (backend.Backend, error) {
	if *backendEndpoint == "" {
		return nil, errors.New("Need to specify --backend_endpoint for when using a Readeck backend")
	}
	if *backendBearerToken == "" {
		return nil, errors.New("Need to specify --backend_endpoint for when using a Readeck backend")
	}
	return backend.NewReadeckConn(*backendEndpoint, *backendBearerToken), nil
}

var allBackends = map[string]backendInit{
	"readeck": initReadeck,
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

func extractBody(r *http.Request) (string, error) {
	buffer := new(strings.Builder)
	if _, err := io.Copy(buffer, r.Body); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func sendPocketRequest(host string, w http.ResponseWriter, r *http.Request, body string) {
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

func getArticles(w http.ResponseWriter, r *http.Request) {
	body, err := extractBody(r)
	if err != nil {
		log.Printf("Errror getting request body: %v", err)
		return
	}
	maybeVLog(r, body)
	sendPocketRequest("getpocket.com", w, r, body)

	w.Write([]byte("{\"test\": \"Get\"}"))
}

func modifyArticles(w http.ResponseWriter, r *http.Request) {
	body, err := extractBody(r)
	if err != nil {
		log.Printf("Errror getting request body: %v", err)
		return
	}
	maybeVLog(r, body)
	sendPocketRequest("getpocket.com", w, r, body)
	w.Write([]byte("{\"test\": \"Modify\"}"))
}

func articleText(w http.ResponseWriter, r *http.Request) {
	body, err := extractBody(r)
	if err != nil {
		log.Printf("Errror getting request body: %v", err)
		return
	}
	maybeVLog(r, body)
	sendPocketRequest("text.getpocket.com", w, r, body)
	/*var req articleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse request body: %v", err.Error()), http.StatusBadRequest)
		return
	}

	var res articleResponse
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, fmt.Sprintf("Unable to serialize response: %v", err.Error()), http.StatusInternalServerError)
		return
	}*/
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got unhandled request at %s", r.URL)
	http.NotFound(w, r)
}

func StartServing() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v3/get", getArticles)
	mux.HandleFunc("/v3/send", modifyArticles)
	mux.HandleFunc("/v3beta/text", articleText)
	mux.HandleFunc("/", catchAll)

	fmt.Printf("Listening on http://localhost:%d\n", *port)

	http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)
}
