// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package gotesting

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	internal "gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

var (
	ciVisibilityTests      = map[*testing.T]civisibility.CiVisibilityTest{}
	ciVisibilityTestsMutex sync.RWMutex
)

type T testing.T

// GetTest is a helper to return *gotesting.T from *testing.T
// internally is just a (*gotesting.T)(t) cast
func GetTest(t *testing.T) *T {
	return (*T)(t)
}

// Run runs f as a subtest of t called name. It runs f in a separate goroutine
// and blocks until f returns or calls t.Parallel to become a parallel test.
// Run reports whether f succeeded (or at least did not fail before calling t.Parallel).
//
// Run may be called simultaneously from multiple goroutines, but all such calls
// must return before the outer test function for t returns.
func (ddt *T) Run(name string, f func(*testing.T)) bool {
	fReflect := reflect.Indirect(reflect.ValueOf(f))
	moduleName, suiteName := utils.GetModuleAndSuiteName(fReflect.Pointer())
	originalFunc := runtime.FuncForPC(fReflect.Pointer())
	// let's increment the test count in the module
	atomic.AddInt32(modulesCounters[moduleName], 1)
	// let's increment the test count in the suite
	atomic.AddInt32(suitesCounters[suiteName], 1)

	t := (*testing.T)(ddt)
	return t.Run(name, func(t *testing.T) {
		module := session.GetOrCreateModuleWithFramework(moduleName, testFramework, runtime.Version())
		suite := module.GetOrCreateSuite(suiteName)
		test := suite.CreateTest(t.Name())
		test.SetTestFunc(originalFunc)
		setCiVisibilityTest(t, test)
		defer func() {
			if r := recover(); r != nil {
				// Panic handling
				test.SetErrorInfo("panic", fmt.Sprint(r), utils.GetStacktrace(2))
				test.Close(civisibility.StatusFail)
				checkModuleAndSuite(module, suite)
				internal.ExitCiVisibility()
				panic(r)
			} else {
				// Normal finalization
				if t.Failed() {
					test.SetTag(ext.Error, true)
					suite.SetTag(ext.Error, true)
					module.SetTag(ext.Error, true)
					test.Close(civisibility.StatusFail)
				} else if t.Skipped() {
					test.Close(civisibility.StatusSkip)
				} else {
					test.Close(civisibility.StatusPass)
				}
				checkModuleAndSuite(module, suite)
			}
		}()

		f(t)
	})
}

// Context returns the CI Visibility context of the Test span
// This may be used to create test's children spans useful for
// integration tests
func (ddt *T) Context() context.Context {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		return ciTest.Context()
	}

	return context.Background()
}

// Failed reports whether the function has failed.
func (ddt *T) Fail() {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fail", "failed test", utils.GetStacktrace(2))
	}

	t.Fail()
}

// FailNow marks the function as having failed and stops its execution
// by calling runtime.Goexit (which then runs all deferred calls in the
// current goroutine).
// Execution will continue at the next test or benchmark.
// FailNow must be called from the goroutine running the
// test or benchmark function, not from other goroutines
// created during the test. Calling FailNow does not stop
// those other goroutines.
func (ddt *T) FailNow() {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("FailNow", "failed test", utils.GetStacktrace(2))
	}

	t.FailNow()
}

// Error is equivalent to Log followed by Fail.
func (ddt *T) Error(args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Error", fmt.Sprint(args...), utils.GetStacktrace(2))
	}

	t.Error(args...)
}

// Errorf is equivalent to Logf followed by Fail.
func (ddt *T) Errorf(format string, args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Errorf", fmt.Sprintf(format, args...), utils.GetStacktrace(2))
	}

	t.Errorf(format, args...)
}

// Fatal is equivalent to Log followed by FailNow.
func (ddt *T) Fatal(args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fatal", fmt.Sprint(args...), utils.GetStacktrace(2))
	}

	t.Fatal(args...)
}

// Fatalf is equivalent to Logf followed by FailNow.
func (ddt *T) Fatalf(format string, args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fatalf", fmt.Sprintf(format, args...), utils.GetStacktrace(2))
	}

	t.Fatalf(format, args...)
}

// Skip is equivalent to Log followed by SkipNow.
func (ddt *T) Skip(args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.CloseWithFinishTimeAndSkipReason(civisibility.StatusSkip, time.Now(), fmt.Sprint(args...))
	}

	t.Skip(args...)
}

// Skipf is equivalent to Logf followed by SkipNow.
func (ddt *T) Skipf(format string, args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.CloseWithFinishTimeAndSkipReason(civisibility.StatusSkip, time.Now(), fmt.Sprintf(format, args...))
	}

	t.Skipf(format, args...)
}

// SkipNow marks the test as having been skipped and stops its execution
// by calling runtime.Goexit.
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// Execution will continue at the next test or benchmark. See also FailNow.
// SkipNow must be called from the goroutine running the test, not from
// other goroutines created during the test. Calling SkipNow does not stop
// those other goroutines.
func (ddt *T) SkipNow() {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.Close(civisibility.StatusSkip)
	}

	t.SkipNow()
}

// Parallel signals that this test is to be run in parallel with (and only with)
// other parallel tests. When a test is run multiple times due to use of
// -test.count or -test.cpu, multiple instances of a single test never run in
// parallel with each other.
func (ddt *T) Parallel() {
	(*testing.T)(ddt).Parallel()
}

// Deadline reports the time at which the test binary will have
// exceeded the timeout specified by the -timeout flag.
//
// The ok result is false if the -timeout flag indicates “no timeout” (0).
func (ddt *T) Deadline() (deadline time.Time, ok bool) {
	return (*testing.T)(ddt).Deadline()
}

// Setenv calls os.Setenv(key, value) and uses Cleanup to
// restore the environment variable to its original value
// after the test.
//
// Because Setenv affects the whole process, it cannot be used
// in parallel tests or tests with parallel ancestors.
func (ddt *T) Setenv(key, value string) {
	(*testing.T)(ddt).Setenv(key, value)
}

func getCiVisibilityTest(t *testing.T) civisibility.CiVisibilityTest {
	ciVisibilityTestsMutex.RLock()
	defer ciVisibilityTestsMutex.RUnlock()

	if v, ok := ciVisibilityTests[t]; ok {
		return v
	}

	return nil
}

func setCiVisibilityTest(t *testing.T, ciTest civisibility.CiVisibilityTest) {
	ciVisibilityTestsMutex.Lock()
	defer ciVisibilityTestsMutex.Unlock()
	ciVisibilityTests[t] = ciTest
}
