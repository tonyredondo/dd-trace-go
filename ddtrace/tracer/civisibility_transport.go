// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package tracer

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"gopkg.in/DataDog/dd-trace-go.v1/internal"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/version"
)

// Constants for CI Visibility API paths and subdomains.
const (
	TestCycleSubdomain = "citestcycle-intake" // Subdomain for test cycle intake.
	TestCyclePath      = "api/v2/citestcycle" // API path for test cycle.
	EvpProxyPath       = "evp_proxy/v2"       // Path for EVP proxy.
)

// Ensure that civisibilityTransport implements the transport interface.
var _ transport = (*civisibilityTransport)(nil)

// civisibilityTransport is a structure that handles sending CI Visibility payloads
// to the Datadog endpoint, either in agentless mode or through the EVP proxy.
type civisibilityTransport struct {
	config           *config           // Configuration for the tracer.
	testCycleUrlPath string            // URL path for the test cycle endpoint.
	client           *http.Client      // HTTP client used to send the requests.
	headers          map[string]string // HTTP headers to be included in the requests.
}

// newCiVisibilityTransport creates and initializes a new civisibilityTransport
// based on the provided tracer configuration. It sets up the appropriate headers
// and determines the URL path based on whether agentless mode is enabled.
//
// Parameters:
//
//	config - The tracer configuration.
//
// Returns:
//
//	A pointer to an initialized civisibilityTransport instance.
func newCiVisibilityTransport(config *config) *civisibilityTransport {
	// Initialize the default headers with encoder metadata.
	defaultHeaders := map[string]string{
		"Datadog-Meta-Lang":             "go",
		"Datadog-Meta-Lang-Version":     strings.TrimPrefix(runtime.Version(), "go"),
		"Datadog-Meta-Lang-Interpreter": runtime.Compiler + "-" + runtime.GOARCH + "-" + runtime.GOOS,
		"Datadog-Meta-Tracer-Version":   version.Tag,
		"Content-Type":                  "application/msgpack",
	}
	if cid := internal.ContainerID(); cid != "" {
		defaultHeaders["Datadog-Container-ID"] = cid
	}
	if eid := internal.EntityID(); eid != "" {
		defaultHeaders["Datadog-Entity-ID"] = eid
	}

	// Determine if agentless mode is enabled through an environment variable.
	agentlessEnabled := internal.BoolEnv(constants.CiVisibilityAgentlessEnabledEnvironmentVariable, false)

	testCycleUrl := ""
	if agentlessEnabled {
		// Agentless mode is enabled.
		defaultHeaders["dd-api-key"] = os.Getenv(constants.ApiKeyEnvironmentVariable)

		// Check for a custom agentless URL.
		agentlessUrl := ""
		if v := os.Getenv(constants.CiVisibilityAgentlessUrlEnvironmentVariable); v != "" {
			agentlessUrl = v
		}

		if agentlessUrl == "" {
			// Use the standard agentless URL format.
			site := "datadoghq.com"
			if v := os.Getenv("DD_SITE"); v != "" {
				site = v
			}

			testCycleUrl = fmt.Sprintf("https://%s.%s/%s", TestCycleSubdomain, site, TestCyclePath)
		} else {
			// Use the custom agentless URL.
			testCycleUrl = fmt.Sprintf("%s/%s", agentlessUrl, TestCyclePath)
		}
	} else {
		// Use agent mode with the EVP proxy.
		defaultHeaders["X-Datadog-EVP-Subdomain"] = TestCycleSubdomain
		testCycleUrl = fmt.Sprintf("%s/%s/%s", config.agentURL.String(), EvpProxyPath, TestCyclePath)
	}

	return &civisibilityTransport{
		config:           config,
		testCycleUrlPath: testCycleUrl,
		client:           config.httpClient,
		headers:          defaultHeaders,
	}
}

// send sends the CI Visibility payload to the Datadog endpoint.
// It prepares the payload, creates the HTTP request, and handles the response.
//
// Parameters:
//
//	p - The payload to be sent.
//
// Returns:
//
//	An io.ReadCloser for reading the response body, and an error if the operation fails.
func (t *civisibilityTransport) send(p *payload) (body io.ReadCloser, err error) {
	ciVisibilityPayload := &civisibilitypayload{p}
	buffer, bufferErr := ciVisibilityPayload.GetBuffer(t.config)
	if bufferErr != nil {
		return nil, fmt.Errorf("cannot create buffer payload: %v", bufferErr)
	}

	// Compress payload
	var gzipBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuffer)
	_, err = io.Copy(gzipWriter, buffer)
	if err != nil {
		return nil, fmt.Errorf("cannot compress request body: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot compress request body: %v", err)
	}

	req, err := http.NewRequest("POST", t.testCycleUrlPath, &gzipBuffer)
	if err != nil {
		return nil, fmt.Errorf("cannot create http request: %v", err)
	}
	for header, value := range t.headers {
		req.Header.Set(header, value)
	}
	req.Header.Set("Content-Length", strconv.Itoa(gzipBuffer.Len()))
	req.Header.Set("Content-Encoding", "gzip")
	response, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	if code := response.StatusCode; code >= 400 {
		// error, check the body for context information and
		// return a nice error.
		msg := make([]byte, 1000)
		n, _ := response.Body.Read(msg)
		_ = response.Body.Close()
		txt := http.StatusText(code)
		if n > 0 {
			return nil, fmt.Errorf("%s (Status: %s)", msg[:n], txt)
		}
		return nil, fmt.Errorf("%s", txt)
	}
	return response.Body, nil
}

// sendStats is a no-op for CI Visibility transport as it does not support sending stats payloads.
//
// Parameters:
//
//	payload - The stats payload to be sent.
//
// Returns:
//
//	An error indicating that stats are not supported.
func (t *civisibilityTransport) sendStats(*statsPayload) error {
	// Stats are not supported by CI Visibility agentless / EVP proxy.
	return nil
}

// endpoint returns the URL path of the test cycle endpoint.
//
// Returns:
//
//	The URL path as a string.
func (t *civisibilityTransport) endpoint() string {
	return t.testCycleUrlPath
}
