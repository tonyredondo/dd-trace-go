// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

//go:generate msgp -unexported -marshal=false -o=civisibility_tslv_msgp.go -tests=false

package tracer

import (
	"strconv"

	"github.com/tinylib/msgp/msgp"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
)

type (
	// ciTestCyclePayloadList implements msgp.Decodable on top of a slice of ciVisibilityPayloads.
	// This type is only used in tests.
	ciTestCyclePayloadList []*ciTestCyclePayload

	ciVisibilityEvents []*ciVisibilityEvent
)

var (
	_ ddtrace.Span   = (*ciVisibilityEvent)(nil)
	_ msgp.Encodable = (*ciVisibilityEvent)(nil)
	_ msgp.Decodable = (*ciVisibilityEvent)(nil)

	_ msgp.Encodable = (*ciVisibilityEvents)(nil)
	_ msgp.Decodable = (*ciVisibilityEvents)(nil)

	_ msgp.Encodable = (*ciTestCyclePayload)(nil)
	_ msgp.Decodable = (*ciTestCyclePayloadList)(nil)
)

type ciTestCyclePayload struct {
	Version  int32                        `msg:"version"`  // version of the payload
	Metadata map[string]map[string]string `msg:"metadata"` // global meta of the payload
	Events   msgp.Raw                     `msg:"events"`   // global meta of the payload
}

type ciVisibilityEvent struct {
	Type    string   `msg:"type"`    // type of civisibility event
	Version int32    `msg:"version"` // version of the type of the event
	Content tslvSpan `msg:"content"` // event content

	span *span `msg:"-"`
}

func (e *ciVisibilityEvent) SetTag(key string, value interface{}) {
	e.span.SetTag(key, value)
	e.Content.Meta = e.span.Meta
	e.Content.Metrics = e.span.Metrics
}

func (e *ciVisibilityEvent) SetOperationName(operationName string) {
	e.span.SetOperationName(operationName)
	e.Content.Name = e.span.Name
}

func (e *ciVisibilityEvent) BaggageItem(key string) string {
	return e.span.BaggageItem(key)
}

func (e *ciVisibilityEvent) SetBaggageItem(key, val string) {
	e.span.SetBaggageItem(key, val)
}

func (e *ciVisibilityEvent) Finish(opts ...ddtrace.FinishOption) {
	e.span.Finish(opts...)
}

func (e *ciVisibilityEvent) Context() ddtrace.SpanContext {
	return e.span.Context()
}

type tslvSpan struct {
	SessionId     uint64             `msg:"test_session_id,omitempty"`    // identifier of this session
	ModuleId      uint64             `msg:"test_module_id,omitempty"`     // identifier of this module
	SuiteId       uint64             `msg:"test_suite_id,omitempty"`      // identifier of this suite
	CorrelationId string             `msg:"itr_correlation_id,omitempty"` // Correlation Id for Intelligent Test Runner transactions
	Name          string             `msg:"name"`                         // operation name
	Service       string             `msg:"service"`                      // service name (i.e. "grpc.server", "http.request")
	Resource      string             `msg:"resource"`                     // resource name (i.e. "/user?id=123", "SELECT * FROM users")
	Type          string             `msg:"type"`                         // protocol associated with the span (i.e. "web", "db", "cache")
	Start         int64              `msg:"start"`                        // span start time expressed in nanoseconds since epoch
	Duration      int64              `msg:"duration"`                     // duration of the span expressed in nanoseconds
	SpanID        uint64             `msg:"span_id,omitempty"`            // identifier of this span
	TraceID       uint64             `msg:"trace_id,omitempty"`           // lower 64-bits of the root span identifier
	ParentID      uint64             `msg:"parent_id,omitempty"`          // identifier of the span's direct parent
	Error         int32              `msg:"error"`                        // error status of the span; 0 means no errors
	Meta          map[string]string  `msg:"meta,omitempty"`               // arbitrary map of metadata
	Metrics       map[string]float64 `msg:"metrics,omitempty"`            // arbitrary map of numeric metrics
}

func getCiVisibilityEvent(span *span) *ciVisibilityEvent {
	switch span.Type {
	case constants.SpanTypeTest:
		return createTestEventFromSpan(span)
	case constants.SpanTypeTestSuite:
		return createTestSuiteEventFromSpan(span)
	case constants.SpanTypeTestModule:
		return createTestModuleEventFromSpan(span)
	case constants.SpanTypeTestSession:
		return createTestSessionEventFromSpan(span)
	default:
		return createTestEventFromSpan(span)
	}
}

func getSpanFromCiVisibilityEvent(civisibilityEvent *ciVisibilityEvent) *span {
	s := civisibilityEvent.span
	if civisibilityEvent.Content.SessionId != 0 {
		s.SetTag(constants.TestSessionIdTagName, civisibilityEvent.Content.SessionId)
	}
	if civisibilityEvent.Content.ModuleId != 0 {
		s.SetTag(constants.TestModuleIdTagName, civisibilityEvent.Content.ModuleId)
	}
	if civisibilityEvent.Content.SuiteId != 0 {
		s.SetTag(constants.TestSuiteIdTagName, civisibilityEvent.Content.SuiteId)
	}
	if civisibilityEvent.Content.CorrelationId != "" {
		s.SetTag(constants.ItrCorrelationIdTagName, civisibilityEvent.Content.CorrelationId)
	}
	return s
}

func createTestEventFromSpan(span *span) *ciVisibilityEvent {
	tSpan := createTslvSpan(span)
	tSpan.SessionId = getAndRemoveMetaToUInt64(span, constants.TestSessionIdTagName)
	tSpan.ModuleId = getAndRemoveMetaToUInt64(span, constants.TestModuleIdTagName)
	tSpan.SuiteId = getAndRemoveMetaToUInt64(span, constants.TestSuiteIdTagName)
	tSpan.CorrelationId = getAndRemoveMeta(span, constants.ItrCorrelationIdTagName)
	tSpan.SpanID = span.SpanID
	tSpan.TraceID = span.TraceID
	return &ciVisibilityEvent{
		span:    span,
		Type:    constants.SpanTypeTest,
		Version: 2,
		Content: tSpan,
	}
}

func createTestSuiteEventFromSpan(span *span) *ciVisibilityEvent {
	tSpan := createTslvSpan(span)
	tSpan.SessionId = getAndRemoveMetaToUInt64(span, constants.TestSessionIdTagName)
	tSpan.ModuleId = getAndRemoveMetaToUInt64(span, constants.TestModuleIdTagName)
	tSpan.SuiteId = getAndRemoveMetaToUInt64(span, constants.TestSuiteIdTagName)
	return &ciVisibilityEvent{
		span:    span,
		Type:    constants.SpanTypeTestSuite,
		Version: 1,
		Content: tSpan,
	}
}

func createTestModuleEventFromSpan(span *span) *ciVisibilityEvent {
	tSpan := createTslvSpan(span)
	tSpan.SessionId = getAndRemoveMetaToUInt64(span, constants.TestSessionIdTagName)
	tSpan.ModuleId = getAndRemoveMetaToUInt64(span, constants.TestModuleIdTagName)
	return &ciVisibilityEvent{
		span:    span,
		Type:    constants.SpanTypeTestModule,
		Version: 1,
		Content: tSpan,
	}
}

func createTestSessionEventFromSpan(span *span) *ciVisibilityEvent {
	tSpan := createTslvSpan(span)
	tSpan.SessionId = getAndRemoveMetaToUInt64(span, constants.TestSessionIdTagName)
	return &ciVisibilityEvent{
		span:    span,
		Type:    constants.SpanTypeTestSession,
		Version: 1,
		Content: tSpan,
	}
}

func createSpanEventFromSpan(span *span) *ciVisibilityEvent {
	tSpan := createTslvSpan(span)
	tSpan.SpanID = span.SpanID
	tSpan.TraceID = span.TraceID
	return &ciVisibilityEvent{
		span:    span,
		Type:    constants.SpanTypeSpan,
		Version: 1,
		Content: tSpan,
	}
}

func createTslvSpan(span *span) tslvSpan {
	return tslvSpan{
		Name:     span.Name,
		Service:  span.Service,
		Resource: span.Resource,
		Type:     span.Type,
		Start:    span.Start,
		Duration: span.Duration,
		ParentID: span.ParentID,
		Error:    span.Error,
		Meta:     span.Meta,
		Metrics:  span.Metrics,
	}
}

func getAndRemoveMeta(span *span, key string) string {
	span.Lock()
	defer span.Unlock()
	if span.Meta == nil {
		span.Meta = make(map[string]string, 1)
	}

	if v, ok := span.Meta[key]; ok {
		delete(span.Meta, key)
		delete(span.Metrics, key)
		return v
	}

	return ""
}

func getAndRemoveMetaToUInt64(span *span, key string) uint64 {
	strValue := getAndRemoveMeta(span, key)
	i, err := strconv.ParseUint(strValue, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
