// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package tracer

import (
	"gopkg.in/DataDog/dd-trace-go.v1/internal/log"
	"sync"
	"time"
)

const (
	// agentlessPayloadMaxLimit is the maximum payload size allowed and should indicate the
	// maximum size of the package that the intake can receive.
	agentlessPayloadMaxLimit = 5 * 1024 * 1024 // 9.5 MB

	// agentlessPayloadSizeLimit specifies the maximum allowed size of the payload before
	// it will trigger a flush to the transport.
	agentlessPayloadSizeLimit = agentlessPayloadMaxLimit / 2
)

var _ traceWriter = (*ciVisibilityTraceWriter)(nil)

type ciVisibilityTraceWriter struct {
	// config holds the tracer configuration
	config *config

	// payload encodes and buffers traces in msgpack format
	payload *civisibilitypayload

	// climit limits the number of concurrent outgoing connections
	climit chan struct{}

	// wg waits for all uploads to finish
	wg sync.WaitGroup
}

func newCiVisibilityTraceWriter(c *config) *ciVisibilityTraceWriter {
	return &ciVisibilityTraceWriter{
		config:  c,
		payload: newCiVisibilityPayload(),
		climit:  make(chan struct{}, concurrentConnectionLimit),
	}
}

func (w *ciVisibilityTraceWriter) add(trace []*span) {
	for _, s := range trace {
		if err := w.payload.push(getCiVisibilityEvent(s)); err != nil {
			log.Error("Error encoding msgpack: %v", err)
		}
		if w.payload.size() > agentlessPayloadSizeLimit {
			w.flush()
		}
	}
}

func (w *ciVisibilityTraceWriter) stop() {
	w.flush()
	w.wg.Wait()
}

func (w *ciVisibilityTraceWriter) flush() {
	if w.payload.itemCount() == 0 {
		return
	}

	w.wg.Add(1)
	w.climit <- struct{}{}
	oldp := w.payload
	w.payload = newCiVisibilityPayload()

	go func(p *civisibilitypayload) {
		defer func(start time.Time) {
			// Once the payload has been used, clear the buffer for garbage
			// collection to avoid a memory leak when references to this object
			// may still be kept by faulty transport implementations or the
			// standard library. See dd-trace-go#976
			p.clear()

			<-w.climit
			w.wg.Done()
		}(time.Now())

		var count, size int
		var err error
		for attempt := 0; attempt <= w.config.sendRetries; attempt++ {
			size, count = p.size(), p.itemCount()
			log.Debug("Sending payload: size: %d traces: %d\n", size, count)
			_, err = w.config.transport.send(&p.payload)
			if err == nil {
				log.Debug("sent traces after %d attempts", attempt+1)
				return
			}
			log.Error("failure sending traces (attempt %d), will retry: %v", attempt+1, err)
			p.reset()
			time.Sleep(time.Millisecond)
		}
		log.Error("lost %d traces: %v", count, err)
	}(oldp)
}
