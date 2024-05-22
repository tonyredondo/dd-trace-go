// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package testing

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"gopkg.in/DataDog/dd-trace-go.v1/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	internal "gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

const (
	spanKind      = "test"
	testFramework = "golang.org/pkg/testing"
)

var (
	session civisibility.CiVisibilityTestSession
	module  civisibility.CiVisibilityTestModule
)

// FinishFunc closes a started span and attaches test status information.
type FinishFunc func()

// Run is a helper function to run a `testing.M` object and gracefully stopping the tracer afterwards
func Run(m *testing.M, opts ...tracer.StartOption) int {
	internal.EnsureCiVisibilityInitialization()
	defer internal.ExitCiVisibility()

	session = civisibility.CreateTestSession()
	module = session.GetOrCreateModuleWithFramework("Package Name", testFramework, runtime.Version())

	var exitCode = m.Run()

	module.Close()
	session.Close(exitCode)

	// Execute test suite
	return exitCode
}

// TB is the minimal interface common to T and B.
type TB interface {
	Failed() bool
	Name() string
	Skipped() bool
}

var _ TB = (*testing.T)(nil)
var _ TB = (*testing.B)(nil)

// StartTest returns a new span with the given testing.TB interface and options. It uses
// tracer.StartSpanFromContext function to start the span with automatically detected information.
func StartTest(tb TB, opts ...Option) (context.Context, FinishFunc) {
	opts = append(opts, WithIncrementSkipFrame())
	return StartTestWithContext(context.Background(), tb, opts...)
}

// StartTestWithContext returns a new span with the given testing.TB interface and options. It uses
// tracer.StartSpanFromContext function to start the span with automatically detected information.
func StartTestWithContext(ctx context.Context, tb TB, opts ...Option) (context.Context, FinishFunc) {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}

	var pc uintptr
	if cfg.originalTestFunc == nil {
		pc, _, _, _ = runtime.Caller(cfg.skip)
	} else {
		pc = reflect.Indirect(reflect.ValueOf(cfg.originalTestFunc)).Pointer()
	}
	suite, _, _, _ := utils.GetPackageAndName(pc)

	ciVisibilitySuite := module.GetOrCreateSuite(suite)
	ciVisibilityTest := ciVisibilitySuite.CreateTest(tb.Name())
	ciVisibilityTest.SetTestFunc(runtime.FuncForPC(pc))

	return ctx, func() {
		var r interface{} = nil

		if r = recover(); r != nil {
			// Panic handling
			ciVisibilityTest.SetErrorInfo("panic", fmt.Sprint(r), utils.GetStacktrace(2))
			ciVisibilityTest.Close(civisibility.StatusFail)
			internal.ExitCiVisibility()
			panic(r)
		} else {
			// Normal finalization
			ciVisibilityTest.SetTag(ext.Error, tb.Failed())

			if tb.Failed() {
				ciVisibilityTest.Close(civisibility.StatusFail)
			} else if tb.Skipped() {
				ciVisibilityTest.Close(civisibility.StatusSkip)
			} else {
				ciVisibilityTest.Close(civisibility.StatusPass)
			}
		}
	}
}
