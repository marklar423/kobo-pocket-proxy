package server_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"proxyserver/readeck"
	"proxyserver/server"
	"strings"
	"testing"

	containers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// This file runs an integration test which spins up a real Readeck server in a container, points
// an instance of the proxy server at it, and runs Pocket-compatible API calls through the proxy
// server.
const readeckImage = "codeberg.org/readeck/readeck:0.19.2"
const readeckUser = "tester"
const readeckPassword = "testpassword"

// A 10x10 blue square PNG.
const testImage = "iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAABHNCSVQICAgIfAhkiAAAABhJREFUGJVjZJjy/z8DEYCJGEWjCqmnEAAnjQKmJi5fSQAAAABJRU5ErkJggg=="

type testServerOptions struct {
	port               int
	backendName        string
	backendEndpoint    string
	backendBearerToken string
}

func (o testServerOptions) Port() int                  { return o.port }
func (testServerOptions) Verbose() bool                { return false }
func (o testServerOptions) BackendName() string        { return o.backendName }
func (o testServerOptions) BackendEndpoint() string    { return o.backendEndpoint }
func (o testServerOptions) BackendBearerToken() string { return o.backendBearerToken }

type readeckEnv struct {
	readeckContainer containers.Container
	readeckBaseUrl   string
	authToken        string
	mockWebsite      *httptest.Server
	proxyBaseUrl     string
}

func (e *readeckEnv) cleanup(t *testing.T) {
	containers.CleanupContainer(t, e.readeckContainer)
	e.mockWebsite.Close()
}

func newReadeckEnv(ctx context.Context, t *testing.T) (*readeckEnv, error) {
	// Create the container
	req := containers.GenericContainerRequest{
		ProviderType: containers.ProviderPodman,
		ContainerRequest: containers.ContainerRequest{
			Image:        readeckImage,
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForExposedPort(),
		},
		Started: true,
	}

	readeckContainer, err := containers.GenericContainer(ctx, req)
	errCleanup := func() { containers.CleanupContainer(t, readeckContainer) }

	if err != nil {
		errCleanup()
		return nil, err
	}
	t.Log("Readeck container started")

	e := &readeckEnv{
		readeckContainer: readeckContainer,
	}

	// Populate the readeck host:port
	host, err := readeckContainer.Host(ctx)
	if err != nil {
		errCleanup()
		return nil, err
	}
	port, err := readeckContainer.MappedPort(ctx, "8000/tcp")
	if err != nil {
		errCleanup()
		return nil, err
	}
	e.readeckBaseUrl = fmt.Sprintf("http://%s:%s", host, port.Port())
	t.Logf("Readeck base URL: %s", e.readeckBaseUrl)

	// Create the initial user
	userCmd := []string{
		"/bin/readeck", "user",
		"-config", "/readeck/config.toml",
		"-email", "test@test.com",
		"-group", "admin",
		"-p", readeckPassword,
		"-u", readeckUser,
	}
	if code, output, err := readeckContainer.Exec(ctx, userCmd); err != nil {
		errCleanup()
		return nil, err
	} else if code != 0 {
		errCleanup()
		return nil, fmt.Errorf("unexpected error from readeck user command: want 0 got %d", code)
	} else {
		buf := new(strings.Builder)
		if _, err := io.Copy(buf, output); err == nil {
			t.Log(buf.String())
		}
	}

	// Get the auth token.
	token, err := readeck.GetAuthToken(e.readeckBaseUrl, "test app", readeckUser, readeckPassword)
	if err != nil {
		errCleanup()
		return nil, err
	}
	e.authToken = token
	t.Logf("Readeck auth token: %s", e.authToken)

	// Start up mock website for Readeck to scrape.
	e.mockWebsite = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.HasSuffix(r.URL.Path, "png") {
			w.Header().Set("Content-Type", "image/png")
			imageData, err := base64.StdEncoding.DecodeString(testImage)
			if err != nil {
				t.Fatalf("Error decoding image base64: %v", err)
			}
			w.Write(imageData)
		}
		if strings.HasSuffix(r.URL.Path, "html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			title := r.URL.Query().Get("title")
			baseUrl := e.mockWebsite.URL

			w.Write([]byte(fmt.Sprintf(`
			<html>
			<head><title>%s</title></head>
			<body>
			  <h1>%s</h1>
			  <div>
			    <p><img src="%s/image1.png" /></p>
			  </div>
				<img src="%s/image2.png" />
			</body>
			</html>
			`, title, title, baseUrl, baseUrl)))
		}
	}))
	t.Logf("Mock website URL: %s", e.mockWebsite.URL)
	errCleanup = func() {
		containers.CleanupContainer(t, readeckContainer)
		e.mockWebsite.Close()
	}

	// Finally, start the proxy server.
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		errCleanup()
		return nil, err
	}
	proxyPort := listener.Addr().(*net.TCPAddr).Port

	// TODO: Does this need manual cleanup?
	go server.StartServing(testServerOptions{
		port:               proxyPort,
		backendName:        "readeck",
		backendEndpoint:    e.readeckBaseUrl,
		backendBearerToken: e.authToken,
	})
	e.proxyBaseUrl = fmt.Sprintf("http://localhost:%d", proxyPort)

	return e, nil
}

func TestServer_ReadeckBackend(t *testing.T) {
	ctx := context.Background()

	env, err := newReadeckEnv(ctx, t)
	if err != nil {
		t.Fatalf("Unexpected error fron newReadeckEnv(): %v", err)
	}
	defer env.cleanup(t)
}
