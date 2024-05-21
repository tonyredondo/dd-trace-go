// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package tracer

import (
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/internal"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/version"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	TestCycleSubdomain = "citestcycle-intake"
	TestCyclePath      = "api/v2/citestcycle"
	EvpProxyPath       = "evp_proxy/v2"
)

var _ transport = (*civisibilityTransport)(nil)

type civisibilityTransport struct {
	config           *config           // config holds the tracer configuration
	testCycleUrlPath string            // the test cycle evp intake path
	client           *http.Client      // the HTTP client used in the POST
	headers          map[string]string // the transport headers
}

func newCiVisibilityTransport(config *config) *civisibilityTransport {
	// initialize the default EncoderPool with Encoder headers
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

	// Check if the agentless environment variable was set.
	agentlessEnabled := internal.BoolEnv("DD_CIVISIBILITY_AGENTLESS_ENABLED", false)

	testCycleUrl := ""
	if agentlessEnabled {
		defaultHeaders["dd-api-key"] = os.Getenv("DD_API_KEY")

		// If agentless is enabled let's check if the custom agentless url environment variable is set
		agentlessUrl := ""
		if v := os.Getenv("DD_CIVISIBILITY_AGENTLESS_URL"); v != "" {
			agentlessUrl = v
		}

		if agentlessUrl == "" {
			// Normal agentless mode

			// Extract the DD_SITE
			site := "datadoghq.com"
			if v := os.Getenv("DD_SITE"); v != "" {
				site = v
			}

			testCycleUrl = fmt.Sprintf("https://%s.%s/%s", TestCycleSubdomain, site, TestCyclePath)
		} else {
			// Agentless mode with custom agentless url
			testCycleUrl = fmt.Sprintf("%s/%s", agentlessUrl, TestCyclePath)
		}
	} else {
		// Agent mode with EvP proxy
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

func (t *civisibilityTransport) send(p *payload) (body io.ReadCloser, err error) {
	ciVisibilityPayload := &civisibilitypayload{p}
	buffer, bufferErr := ciVisibilityPayload.GetBuffer(t.config)
	if bufferErr != nil {
		return nil, fmt.Errorf("cannot create buffer payload: %v", bufferErr)
	}

	req, err := http.NewRequest("POST", t.testCycleUrlPath, buffer)
	if err != nil {
		return nil, fmt.Errorf("cannot create http request: %v", err)
	}
	for header, value := range t.headers {
		req.Header.Set(header, value)
	}
	req.Header.Set("Content-Length", strconv.Itoa(p.size()))
	response, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	if code := response.StatusCode; code >= 400 {
		// error, check the body for context information and
		// return a nice error.
		msg := make([]byte, 1000)
		n, _ := response.Body.Read(msg)
		response.Body.Close()
		txt := http.StatusText(code)
		if n > 0 {
			return nil, fmt.Errorf("%s (Status: %s)", msg[:n], txt)
		}
		return nil, fmt.Errorf("%s", txt)
	}
	return response.Body, nil
}

func (t *civisibilityTransport) sendStats(s *statsPayload) error {
	// Stats is not supported by CI Visibility agentless / evp proxy
	return nil
}

func (t *civisibilityTransport) endpoint() string {
	return t.testCycleUrlPath
}
