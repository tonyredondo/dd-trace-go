// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package tracer

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinylib/msgp/msgp"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
)

func TestCiVisibilityTransport(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		payload [][]*span
	}{
		{getTestTrace(1, 1)},
		{getTestTrace(10, 1)},
		{getTestTrace(100, 10)},
	}

	remainingEvents := 1000 + 10 + 1
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		metaLang := r.Header.Get("Datadog-Meta-Lang")
		assert.NotNil(metaLang)

		apikey := r.Header.Get("dd-api-key")
		assert.Equal("12345", apikey)

		contentType := r.Header.Get("Content-Type")
		assert.Equal("application/msgpack", contentType)

		assert.True(strings.HasSuffix(r.RequestURI, TestCyclePath))

		bodyBuffer := new(bytes.Buffer)
		_, err := bodyBuffer.ReadFrom(r.Body)
		assert.NoError(err)

		var testCyclePayload ciTestCyclePayload
		err = msgp.Decode(bodyBuffer, &testCyclePayload)
		assert.NoError(err)

		var events ciVisibilityEvents
		err = msgp.Decode(bytes.NewBuffer(testCyclePayload.Events), &events)
		assert.NoError(err)

		remainingEvents = remainingEvents - len(events)
	}))
	defer srv.Close()
	c := config{
		ciVisibilityEnabled: true,
		httpClient:          defaultHTTPClient(0),
	}

	// Set CI Visibility environment variables for the test
	t.Setenv(constants.CiVisibilityAgentlessEnabledEnvironmentVariable, "1")
	t.Setenv(constants.CiVisibilityAgentlessUrlEnvironmentVariable, srv.URL)
	t.Setenv(constants.ApiKeyEnvironmentVariable, "12345")

	for _, tc := range testCases {
		transport := newCiVisibilityTransport(&c)

		p := newCiVisibilityPayload()
		for _, t := range tc.payload {
			for _, span := range t {
				err := p.push(getCiVisibilityEvent(span))
				assert.NoError(err)
			}
		}

		_, err := transport.send(p.payload)
		assert.NoError(err)
	}
	assert.Equal(hits, len(testCases))
	assert.Equal(remainingEvents, 0)
}
