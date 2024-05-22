// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package constants

const (
	// TestModule indicates the test module name
	TestModule = "test.module"

	// TestSuite indicates the test suite name.
	TestSuite = "test.suite"

	// TestName indicates the test name.
	TestName = "test.name"

	// TestType indicates the type of the test (test, benchmark).
	TestType = "test.type"

	// TestFramework indicates the test framework name.
	TestFramework = "test.framework"

	// TestFrameworkVersion indicates the test framework version.
	TestFrameworkVersion = "test.framework_version"

	// TestStatus indicates the test execution status.
	TestStatus = "test.status"

	// TestSkipReason indicates the skip reason of the test.
	TestSkipReason = "test.skip_reason"

	// TestSourceFile indicates the source file where the test is located.
	TestSourceFile = "test.source.file"

	// TestSourceStartLine indicates the line of the source file where the test starts.
	TestSourceStartLine = "test.source.start"

	// TestCodeOwners indicates the test codeowners.
	TestCodeOwners = "test.codeowners"

	// TestCommand indicates the test command.
	TestCommand = "test.command"

	// TestCommandExitCode indicates the test command exit code.
	TestCommandExitCode = "test.exit_code"

	// TestCommandWorkingDirectory indicates the test command working directory relative to the source root.
	TestCommandWorkingDirectory = "test.working_directory"
)

// Define valid test status types.
const (
	// TestStatusPass marks test execution as passed.
	TestStatusPass = "pass"

	// TestStatusFail marks test execution as failed.
	TestStatusFail = "fail"

	// TestStatusSkip marks test execution as skipped.
	TestStatusSkip = "skip"
)

// Define valid test types.
const (
	// TestTypeTest defines test type as test.
	TestTypeTest = "test"

	// TestTypeBenchmark defines test type as benchmark.
	TestTypeBenchmark = "benchmark"
)
