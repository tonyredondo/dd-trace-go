// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package civisibility

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
)

// Test Suite

var _ CiVisibilityTestSuite = (*tslvTestSuite)(nil)

type tslvTestSuite struct {
	ciVisibilityCommon
	module  *tslvTestModule
	suiteId uint64
	name    string
}

func createTestSuite(module *tslvTestModule, name string, startTime time.Time) CiVisibilityTestSuite {
	if module == nil {
		return nil
	}

	operationName := "test_suite"
	if module.framework != "" {
		operationName = fmt.Sprintf("%s.%s", strings.ToLower(module.framework), operationName)
	}

	resourceName := name

	// suite tags should include also the module and session tags so the backend can calculate the module and session fingerprint from the suite
	suiteTags := append(module.tags, tracer.Tag(constants.TestSuite, name))
	testOpts := append(fillCommonTags([]tracer.StartSpanOption{
		tracer.ResourceName(resourceName),
		tracer.SpanType(constants.SpanTypeTestSuite),
		tracer.StartTime(startTime),
	}), suiteTags...)

	span, ctx := tracer.StartSpanFromContext(context.Background(), operationName, testOpts...)
	suiteId := span.Context().SpanID()
	if module.session != nil {
		span.SetTag(constants.TestSessionIdTagName, fmt.Sprint(module.session.sessionId))
	}
	span.SetTag(constants.TestModuleIdTagName, fmt.Sprint(module.moduleId))
	span.SetTag(constants.TestSuiteIdTagName, fmt.Sprint(suiteId))

	suite := &tslvTestSuite{
		module:  module,
		suiteId: suiteId,
		name:    name,
		ciVisibilityCommon: ciVisibilityCommon{
			startTime: startTime,
			tags:      suiteTags,
			span:      span,
			ctx:       ctx,
		},
	}

	// We need to ensure to close everything before ci visibility is exiting.
	// In ci visibility mode we try to never lose data
	pushCiVisibilityCloseAction(func() { suite.Close() })

	return suite
}

func (t *tslvTestSuite) Context() context.Context       { return t.ctx }
func (t *tslvTestSuite) StartTime() time.Time           { return t.startTime }
func (t *tslvTestSuite) Name() string                   { return t.name }
func (t *tslvTestSuite) Module() CiVisibilityTestModule { return t.module }
func (t *tslvTestSuite) SetError(err error)             { t.span.SetTag(ext.Error, err) }
func (t *tslvTestSuite) SetErrorInfo(errType string, message string, callstack string) {
	// set the span with error:1
	t.span.SetTag(ext.Error, true)

	// set the error type
	if errType != "" {
		t.span.SetTag(ext.ErrorType, errType)
	}

	// set the error message
	if message != "" {
		t.span.SetTag(ext.ErrorMsg, message)
	}

	// set the error stacktrace
	if callstack != "" {
		t.span.SetTag(ext.ErrorStack, callstack)
	}
}
func (t *tslvTestSuite) SetTag(key string, value interface{}) { t.span.SetTag(key, value) }
func (t *tslvTestSuite) Close()                               { t.CloseWithFinishTime(time.Now()) }
func (t *tslvTestSuite) CloseWithFinishTime(finishTime time.Time) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return
	}

	t.span.Finish(tracer.FinishTime(finishTime))
	t.closed = true
}
func (t *tslvTestSuite) CreateTest(name string) CiVisibilityTest {
	return t.CreateTestWithStartTime(name, time.Now())
}
func (t *tslvTestSuite) CreateTestWithStartTime(name string, startTime time.Time) CiVisibilityTest {
	return createTest(t, name, startTime)
}
