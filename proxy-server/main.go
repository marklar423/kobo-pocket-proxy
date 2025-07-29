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
