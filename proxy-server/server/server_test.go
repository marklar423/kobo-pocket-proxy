package server_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"proxyserver/pocketapi"
	"proxyserver/readeck"
	"proxyserver/server"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	toml "github.com/pelletier/go-toml"
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

func postJSON[T any, U any](t *testing.T, url string, requestBody T, responseBody *U) error {
	t.Logf("Sending POST to %s", url)

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(requestBody); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, &buffer)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
		if err != nil {
			return fmt.Errorf("failed to parse response JSON: %w", err)
		}
	}

	return nil
}

func findUnusedPort() (int, error) {
	// This is a hack that starts listening on an unused port and immediately closes
	// it. It'll work most of the time, unless some other process grabs it right after.
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

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
	// Modify the Readeck config to allow extracting loopback addresses.
	config, err := toml.LoadBytes(nil)
	if err != nil {
		return nil, err
	}
	config.Set("extractor.denied_ips", []string{"192.168.0.0/128"})
	newConfig, err := config.ToTomlString()
	if err != nil {
		return nil, err
	}

	// Create the container
	req := containers.GenericContainerRequest{
		ProviderType: containers.ProviderPodman,
		ContainerRequest: containers.ContainerRequest{
			Image:        readeckImage,
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForExposedPort(),
			Files: []containers.ContainerFile{{
				ContainerFilePath: "/readeck/config.toml",
				Reader:            strings.NewReader(newConfig),
				FileMode:          0o777,
			}},
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
		t.Logf("Mock website got request at %s", r.URL.String())
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

			fmt.Fprintf(w, `
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
			`, title, title, baseUrl, baseUrl)
		}
	}))
	t.Logf("Mock website URL: %s", e.mockWebsite.URL)
	errCleanup = func() {
		containers.CleanupContainer(t, readeckContainer)
		e.mockWebsite.Close()
	}

	// Finally, start the proxy server.
	proxyPort, err := findUnusedPort()
	if err != nil {
		errCleanup()
		return nil, err
	}

	// TODO: Does this need manual cleanup?
	go server.StartServing(testServerOptions{
		port:               proxyPort,
		backendName:        "readeck",
		backendEndpoint:    e.readeckBaseUrl,
		backendBearerToken: e.authToken,
	})
	e.proxyBaseUrl = fmt.Sprintf("http://localhost:%d", proxyPort)

	// Wait a little to give the server time to start.
	// This might be flaky.
	time.Sleep(2 * time.Second)

	return e, nil
}

func TestServer_ReadeckBackend(t *testing.T) {
	ctx := context.Background()

	env, err := newReadeckEnv(ctx, t)
	if err != nil {
		t.Fatalf("Unexpected error fron newReadeckEnv(): %v", err)
	}
	defer env.cleanup(t)

	t.Run("InitialState", func(t *testing.T) {
		req := pocketapi.GetRequest{ContentType: "article", DetailType: "complete", State: "all"}
		var res pocketapi.GetResponse
		if err := postJSON(t, fmt.Sprintf("%s/v3/get", env.proxyBaseUrl), req, &res); err != nil {
			t.Fatalf("Unexpected error calling /v3/get: %v", err)
		}
		if len(res.List) > 0 {
			t.Errorf("Unexpected number of items, want 0 got %d", len(res.List))
		}
	})

	t.Run("SaveArticle", func(t *testing.T) {
		// Save an article
		articleUrl := fmt.Sprintf("%s/test1.html?title=Title1", env.mockWebsite.URL)
		req := pocketapi.SendRequest{
			Actions: []pocketapi.SendAction{
				{Action: "add", URL: articleUrl},
			},
		}
		var res pocketapi.SendResponse
		if err := postJSON(t, fmt.Sprintf("%s/v3/send", env.proxyBaseUrl), req, &res); err != nil {
			t.Fatalf("Unexpected error calling /v3/send: %v", err)
		}
		if !res.ActionResults[0] {
			t.Errorf("Unexpected response from send action result: want true got false. Error: %v", res.ActionErrors[0])
		}

		// Wait a little for the article to download (might be flaky).
		time.Sleep(time.Second)

		// Article exists
		getReq := pocketapi.GetRequest{ContentType: "article", DetailType: "complete", State: "all"}
		var getRes pocketapi.GetResponse
		if err := postJSON(t, fmt.Sprintf("%s/v3/get", env.proxyBaseUrl), getReq, &getRes); err != nil {
			t.Fatalf("Unexpected error calling /v3/get: %v", err)
		}
		t.Logf("Get response: %+v", getRes)

		if len(getRes.List) != 1 {
			t.Errorf("Unexpected number of items, want 1 got %d", len(getRes.List))
		}
		for k, v := range getRes.List {
			if v.ItemID != k {
				t.Errorf("Get map key doesn't match, key %s != item ID %s", k, v.ItemID)
			}
			if v.GivenURL != articleUrl || v.ResolvedURL != articleUrl {
				t.Errorf("Unexpected given or resolved URL, want %s got %s and %s", articleUrl, v.GivenURL, v.ResolvedURL)
			}
			if v.GivenTitle != "Title1" || v.ResolvedTitle != "Title1" {
				t.Errorf("Unexpected given or resolved title, want Title1 got %s and %s", v.GivenTitle, v.ResolvedTitle)
			}
			if len(v.Images) != 2 {
				t.Errorf("Unexpected number of articles images, want 2 got %d", len(v.Images))
			}
			wantImages := []string{
				fmt.Sprintf("%s/image1.png", env.mockWebsite.URL),
				fmt.Sprintf("%s/image2.png", env.mockWebsite.URL),
			}
			var gotImages []string
			for _, img := range v.Images {
				gotImages = append(gotImages, img.Src)
				if img.Height != "10" || img.Width != "10" {
					t.Errorf("Unexpected image dimensions, want 10x10, got %sx%s", img.Width, img.Height)
				}
			}
			if diff := cmp.Diff(wantImages, gotImages, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Article images mismatch (-want +got):\n%s", diff)
			}
		}
	})
}
