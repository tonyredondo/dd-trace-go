// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package civisibility

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

const (
	spanKind      = "test"
	testFramework = "golang.org/pkg/testing"
)

type (
	// civisibilityCloseAction action to be executed when ci visibility is closing
	civisibilityCloseAction func()
)

var (
	// ciVisibilityInitializationOnce ensure we initialize the ci visibility tracer only once
	ciVisibilityInitializationOnce sync.Once

	// closeActions ci visibility close actions
	closeActions []civisibilityCloseAction

	// closeActionsMutex ci visibility close actions mutex
	closeActionsMutex sync.Mutex

	session CiVisibilityTestSession
	module  CiVisibilityTestModule
	suites  = map[string]CiVisibilityTestSuite{}
)

func ensureCiVisibilityInitialization() {
	ciVisibilityInitializationOnce.Do(func() {
		// Preload all CI and Git tags.
		ciTags := utils.GetCiTags()

		// Check if DD_SERVICE has been set; otherwise we default to repo name.
		var opts []tracer.StartOption
		if v := os.Getenv("DD_SERVICE"); v == "" {
			if repoUrl, ok := ciTags[constants.GitRepositoryURL]; ok {
				// regex to sanitize the repository url to be used as a service name
				repoRegex := regexp.MustCompile(`(?m)/([a-zA-Z0-9\\\-_.]*)$`)
				matches := repoRegex.FindStringSubmatch(repoUrl)
				if len(matches) > 1 {
					repoUrl = strings.TrimSuffix(matches[1], ".git")
				}
				opts = append(opts, tracer.WithService(repoUrl))
			}
		}

		// Initialize tracer
		tracer.Start(opts...)

		// Handle SIGINT and SIGTERM
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signals
			exitCiVisibility()
			os.Exit(1)
		}()
	})
}

func pushCiVisibilityCloseAction(action civisibilityCloseAction) {
	closeActionsMutex.Lock()
	defer closeActionsMutex.Unlock()
	closeActions = append([]civisibilityCloseAction{action}, closeActions...)
}

func exitCiVisibility() {
	closeActionsMutex.Lock()
	defer closeActionsMutex.Unlock()
	for _, v := range closeActions {
		v()
	}

	tracer.Flush()
	tracer.Stop()
}

// FinishFunc closes a started span and attaches test status information.
type FinishFunc func()

// Run is a helper function to run a `testing.M` object and gracefully stopping the tracer afterwards
func Run(m *testing.M, opts ...tracer.StartOption) int {
	ensureCiVisibilityInitialization()
	defer exitCiVisibility()

	session = CreateTestSession()
	module = session.GetOrCreateModule("Package Name")

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
			ciVisibilityTest.Close(StatusFail)
			exitCiVisibility()
			panic(r)
		} else {
			// Normal finalization
			ciVisibilityTest.SetTag(ext.Error, tb.Failed())

			if tb.Failed() {
				ciVisibilityTest.Close(StatusFail)
			} else if tb.Skipped() {
				ciVisibilityTest.Close(StatusSkip)
			} else {
				ciVisibilityTest.Close(StatusPass)
			}
		}
	}
}
