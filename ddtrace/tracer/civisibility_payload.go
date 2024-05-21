// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package tracer

import (
	"bytes"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/globalconfig"
	"sync/atomic"

	"github.com/tinylib/msgp/msgp"
)

type civisibilitypayload struct {
	payload
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
	return &civisibilitypayload{*newPayload()}
}

// Get complete civisibility payload
func (p *civisibilitypayload) GetBuffer() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(&p.payload)
	if err != nil {
		return nil, err
	}

	visibilityPayload := ciTestCyclePayload{
		Version: 1,
		Metadata: map[string]map[string]string{
			"*": {
				"language":   "go",
				"runtime-id": globalconfig.RuntimeID(),
			},
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
