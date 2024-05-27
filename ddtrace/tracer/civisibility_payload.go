// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package tracer

import (
	"bytes"
	"sync/atomic"

	"github.com/tinylib/msgp/msgp"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/globalconfig"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/version"
)

type civisibilitypayload struct {
	*payload
}

// push pushes a new item into the stream.
func (p *civisibilitypayload) push(event *ciVisibilityEvent) error {
	p.buf.Grow(event.Msgsize())
	if err := msgp.Encode(&p.buf, event); err != nil {
		return err
	}
	atomic.AddUint32(&p.count, 1)
	p.updateHeader()
	return nil
}

func newCiVisibilityPayload() *civisibilitypayload {
	return &civisibilitypayload{newPayload()}
}

// GetBuffer gets the complete body of the CiVisibility payload
func (p *civisibilitypayload) GetBuffer(config *config) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(p.payload)
	if err != nil {
		return nil, err
	}

	allMetadata := map[string]string{
		"language":        "go",
		"runtime-id":      globalconfig.RuntimeID(),
		"library_version": version.Tag,
	}

	if config.env != "" {
		allMetadata["env"] = config.env
	}

	visibilityPayload := ciTestCyclePayload{
		Version: 1,
		Metadata: map[string]map[string]string{
			"*": allMetadata,
		},
		Events: buf.Bytes(),
	}

	buf = new(bytes.Buffer)
	err = msgp.Encode(buf, &visibilityPayload)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
