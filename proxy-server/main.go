package main

import (
	"flag"
	"proxyserver/server"
)

var port = flag.Int("port", 8080, "HTTP port to listen on")
var verbose = flag.Bool("verbost", true, "If true, dumps all request fields to stdout")
var backendName = flag.String("backend", "readeck", "The name of the backend to forward API calls to")
var backendEndpoint = flag.String("backend_endpoint", "", "The backend API endpoint")
var backendBearerToken = flag.String("backend_bearer_token", "", "The backend API bearer token used for authentication")

type FlagOptions struct{}

func (FlagOptions) Port() int                  { return *port }
func (FlagOptions) Verbose() bool              { return *verbose }
func (FlagOptions) BackendName() string        { return *backendName }
func (FlagOptions) BackendEndpoint() string    { return *backendEndpoint }
func (FlagOptions) BackendBearerToken() string { return *backendBearerToken }

func main() {
	flag.Parse()
	server.StartServing(FlagOptions{})
}
