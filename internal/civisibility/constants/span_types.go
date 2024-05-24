// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package constants

const (
	// SpanTypeTest marks a span as a test execution.
	SpanTypeTest = "test"

	// SpanTypeBenchmark marks a span as a benchmark execution.
	SpanTypeBenchmark = "benchmark"

	// SpanTypeTestSuite marks a span as a test suite
	SpanTypeTestSuite = "test_suite_end"

	// SpanTypeTestModule marks a span as a test module
	SpanTypeTestModule = "test_module_end"

	// SpanTypeTestSession marks a span as a test session
	SpanTypeTestSession = "test_session_end"

	// SpanTypeSpan marks a span as a span event
	SpanTypeSpan = "span"
)
