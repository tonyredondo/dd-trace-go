// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package mock

import (
	"os"
	"testing"
	"time"
	_ "unsafe"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
	"gopkg.in/DataDog/dd-trace-go.v1/civisibility/gotesting"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
)

var tracer mocktracer.Tracer

type SSpan struct {
	Name       string
	Tags       map[string]any
	FinishTime time.Time
	StartTime  time.Time
	SpanId     uint64
	TraceId    uint64
	ParentId   uint64
}

func TestMain(om *testing.M) {
	// Initialize civisibility using the mocktracer for testing
	tracer = civisibility.InitializeCiVisibilityMock()

	// Run tests using the mocked version of CI Visibility
	m := (*gotesting.M)(om)
	os.Exit(m.Run())
}

func TestDummyTestForMocking(ot *testing.T) {
	t := (*gotesting.T)(ot)
	t.Run("Child", func(ot *testing.T) {
		t := (*gotesting.T)(ot)
		if span, ok := ddtracer.SpanFromContext(t.Context()); ok {
			span.SetTag("Custom Tag", "Custom Value")
		}

		span, _ := ddtracer.StartSpanFromContext(t.Context(), "Custom Span")
		defer span.Finish()
		span.SetTag("Key", "Value")
	})
}

// TestAssertMock ensures that the mock spans are correct.
func TestAssertMock(t *testing.T) {

	// Retrieve the finished spans from the mock tracer
	spans := tracer.FinishedSpans()

	// Fail the test if no spans are found
	if len(spans) == 0 {
		t.Error("No mock spans found")
	}

	// Convert the spans to a serializable structure
	serializableSpans := make([]SSpan, len(spans))
	for i, span := range spans {
		serializableSpans[i] = SSpan{
			Name:       span.OperationName(),
			Tags:       span.Tags(),
			FinishTime: span.FinishTime(),
			StartTime:  span.StartTime(),
			SpanId:     span.SpanID(),
			TraceId:    span.TraceID(),
			ParentId:   span.ParentID(),
		}
	}

	// Match the serializable spans with an expected snapshot, ignoring certain dynamic fields
	snaps.MatchJSON(t, serializableSpans,
		match.Any(
			"#.StartTime",
			"#.FinishTime",
			"#.Tags.ci\\.workspace_path",
			"#.Tags.git\\.branch",
			"#.Tags.git\\.commit\\.author\\.date",
			"#.Tags.git\\.commit\\.author\\.email",
			"#.Tags.git\\.commit\\.author\\.name",
			"#.Tags.git\\.commit\\.committer\\.date",
			"#.Tags.git\\.commit\\.committer\\.email",
			"#.Tags.git\\.commit\\.committer\\.name",
			"#.Tags.git\\.commit\\.message",
			"#.Tags.git\\.commit\\.sha",
			"#.Tags.git\\.repository_url",
			"#.Tags.os\\.architecture",
			"#.Tags.os\\.platform",
			"#.Tags.os\\.version",
			"#.Tags.runtime\\.version",
			"#.Tags.test\\.command",
			"#.Tags.test\\.framework_version",
			"#.Tags.test\\.source\\.start"))
}
