// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package civisibility

import (
	"context"
	"runtime"
	"sync"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

type TestResultStatus int

const (
	StatusPass TestResultStatus = 0
	StatusFail                  = 1
	StatusSkip                  = 2
)

type ciVisibilityTslvEvent interface {
	Context() context.Context
	StartTime() time.Time
	SetError(err error)
	SetErrorInfo(errType string, message string, callstack string)
	SetTag(key string, value interface{})
}

type CiVisibilityTestSession interface {
	ciVisibilityTslvEvent
	Command() string
	Framework() string
	WorkingDirectory() string
	Close(exitCode int)
	CloseWithFinishTime(exitCode int, finishTime time.Time)
	GetOrCreateModule(name string) CiVisibilityTestModule
	GetOrCreateModuleWithFramework(name string, framework string, frameworkVersion string) CiVisibilityTestModule
	GetOrCreateModuleWithFrameworkAndStartTime(name string, framework string, frameworkVersion string, startTime time.Time) CiVisibilityTestModule
}

type CiVisibilityTestModule interface {
	ciVisibilityTslvEvent
	Session() CiVisibilityTestSession
	Framework() string
	Name() string
	Close()
	CloseWithFinishTime(finishTime time.Time)
	GetOrCreateSuite(name string) CiVisibilityTestSuite
	GetOrCreateSuiteWithStartTime(name string, startTime time.Time) CiVisibilityTestSuite
}

type CiVisibilityTestSuite interface {
	ciVisibilityTslvEvent
	Module() CiVisibilityTestModule
	Name() string
	Close()
	CloseWithFinishTime(finishTime time.Time)
	CreateTest(name string) CiVisibilityTest
	CreateTestWithStartTime(name string, startTime time.Time) CiVisibilityTest
}

type CiVisibilityTest interface {
	ciVisibilityTslvEvent
	Name() string
	Suite() CiVisibilityTestSuite
	Close(status TestResultStatus)
	CloseWithFinishTime(status TestResultStatus, finishTime time.Time)
	CloseWithFinishTimeAndSkipReason(status TestResultStatus, finishTime time.Time, skipReason string)
	SetTestFunc(fn *runtime.Func)
	SetBenchmarkData(measureType string, data map[string]any)
}

// common
var _ ciVisibilityTslvEvent = (*ciVisibilityCommon)(nil)

type ciVisibilityCommon struct {
	startTime time.Time

	tags   []tracer.StartSpanOption
	span   tracer.Span
	ctx    context.Context
	mutex  sync.Mutex
	closed bool
}

func (c *ciVisibilityCommon) Context() context.Context { return c.ctx }
func (c *ciVisibilityCommon) StartTime() time.Time     { return c.startTime }
func (c *ciVisibilityCommon) SetError(err error) {
	c.span.SetTag(ext.Error, err)
}
func (c *ciVisibilityCommon) SetErrorInfo(errType string, message string, callstack string) {
	// set the span with error:1
	c.span.SetTag(ext.Error, true)

	// set the error type
	if errType != "" {
		c.span.SetTag(ext.ErrorType, errType)
	}

	// set the error message
	if message != "" {
		c.span.SetTag(ext.ErrorMsg, message)
	}

	// set the error stacktrace
	if callstack != "" {
		c.span.SetTag(ext.ErrorStack, callstack)
	}
}
func (c *ciVisibilityCommon) SetTag(key string, value interface{}) { c.span.SetTag(key, value) }

func fillCommonTags(opts []tracer.StartSpanOption) []tracer.StartSpanOption {
	opts = append(opts, []tracer.StartSpanOption{
		tracer.Tag(constants.Origin, constants.CIAppTestOrigin),
		tracer.Tag(ext.ManualKeep, true),
	}...)

	// Apply CI tags
	for k, v := range utils.GetCiTags() {
		opts = append(opts, tracer.Tag(k, v))
	}

	return opts
}
