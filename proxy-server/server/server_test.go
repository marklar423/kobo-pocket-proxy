package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"proxyserver/pocketapi"
	"proxyserver/readeck"
	"proxyserver/server"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	toml "github.com/pelletier/go-toml"
	containers "github.com/testcontainers/testcontainers-go"
	containernet "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// This file runs an integration test which spins up a real Readeck server in a container, points
// an instance of the proxy server at it, and runs Pocket-compatible API calls through the proxy
// server.
const readeckImage = "codeberg.org/readeck/readeck:0.19.2"
const readeckUser = "tester"
const readeckPassword = "testpassword"

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

func extractContainerUrl(ctx context.Context, c containers.Container, mappedPort nat.Port) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.MappedPort(ctx, mappedPort)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil

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
	network            *containers.DockerNetwork
	readeckContainer   containers.Container
	readeckBaseUrl     string
	authToken          string
	mockWebsite        containers.Container
	mockWebsiteBaseUrl string
	proxyBaseUrl       string
}

func (e *readeckEnv) cleanup(t *testing.T) {
	containers.CleanupNetwork(t, e.network)
	containers.CleanupContainer(t, e.readeckContainer)
	containers.CleanupContainer(t, e.mockWebsite)
}

func newReadeckEnv(ctx context.Context, t *testing.T) (*readeckEnv, error) {
	e := &readeckEnv{}

	containerNet, err := containernet.New(ctx)
	if err != nil {
		return nil, err
	}
	e.network = containerNet
	errCleanup := func() { e.cleanup(t) }

	// Start up mock website for Readeck to scrape.
	mockWebsite, err := containers.GenericContainer(ctx, containers.GenericContainerRequest{
		ContainerRequest: containers.ContainerRequest{
			ExposedPorts:   []string{"9090/tcp"},
			Networks:       []string{e.network.Name},
			NetworkAliases: map[string][]string{e.network.Name: {"mockwebsite"}},
			// Note: the mockwebsite container must be built before running this.
			// Ideally this would be done automatically with `FromDockerfile`, but
			// I couldn't get it working (kept complaining about insufficient UIDs).
			Image: "localhost/mockwebsite:latest",
		},
		Started: true,
	})
	if err != nil {
		errCleanup()
		return nil, err
	}

	e.mockWebsite = mockWebsite

	cip, _ := e.mockWebsite.ContainerIP(ctx)
	t.Logf("Mockwebsite IP: %s", cip)
	mockWebsiteBaseUrl, err := extractContainerUrl(ctx, e.mockWebsite, "9090/tcp")
	if err != nil {
		errCleanup()
		return nil, err
	}
	e.mockWebsiteBaseUrl = mockWebsiteBaseUrl
	t.Logf("Mock website base URL: %s", e.mockWebsiteBaseUrl)

	// Modify the Readeck config to allow extracting loopback addresses.
	config, err := toml.LoadBytes(nil)
	if err != nil {
		errCleanup()
		return nil, err
	}
	config.Set("extractor.denied_ips", []string{"192.168.0.0/128"})
	newConfig, err := config.ToTomlString()
	if err != nil {
		errCleanup()
		return nil, err
	}

	// Create the Readeck container
	readeckContainer, err := containers.GenericContainer(ctx, containers.GenericContainerRequest{
		ContainerRequest: containers.ContainerRequest{
			Image:        readeckImage,
			ExposedPorts: []string{"8000/tcp"},
			Networks:     []string{e.network.Name},
			WaitingFor:   wait.ForExposedPort(),
			Files: []containers.ContainerFile{{
				ContainerFilePath: "/readeck/config.toml",
				Reader:            strings.NewReader(newConfig),
				FileMode:          0o777,
			}},
		},
		Started: true,
	})
	if err != nil {
		errCleanup()
		return nil, err
	}

	t.Log("Readeck container started")

	e.readeckContainer = readeckContainer

	// Populate the readeck host:port
	readeckBaseUrl, err := extractContainerUrl(ctx, e.readeckContainer, "8000/tcp")
	if err != nil {
		errCleanup()
		return nil, err
	}
	e.readeckBaseUrl = readeckBaseUrl
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
		articleUrl := fmt.Sprintf("%s/test1.html?title=Title1", env.mockWebsiteBaseUrl)
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
				fmt.Sprintf("%s/image1.png", env.mockWebsiteBaseUrl),
				fmt.Sprintf("%s/image2.png", env.mockWebsiteBaseUrl),
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
