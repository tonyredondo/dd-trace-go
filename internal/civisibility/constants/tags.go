// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package constants

const (
	// Origin tag
	Origin = "_dd.origin"

	// CIAppTestOrigin defines the CIApp test origin value
	CIAppTestOrigin = "ciapp-test"

	// TestSessionIdTagName defines the test session id tag name for the CI Visibility Protocol
	TestSessionIdTagName string = "test_session_id"

	// TestModuleIdTagName defines the test module id tag name for the CI Visibility Protocol
	TestModuleIdTagName string = "test_module_id"

	// TestSuiteIdTagName defines the test suite id tag name for the CI Visibility Protocol
	TestSuiteIdTagName string = "test_suite_id"

	// ItrCorrelationIdTagName defines the correlation id for the intelligent test runner tag name for the CI Visibility Protocol
	ItrCorrelationIdTagName string = "itr_correlation_id"
)

// Coverage tags
const (
	// CodeCoverageEnabledTagName defines if the code coverage has been enabled
	CodeCoverageEnabledTagName string = "test.code_coverage.enabled"

	// CodeCoveragePercentageOfTotalLines defines the percentage of total code coverage by a session and module
	CodeCoveragePercentageOfTotalLines string = "test.code_coverage.lines_pct"
)
