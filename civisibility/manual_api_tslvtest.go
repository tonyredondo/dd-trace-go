// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package civisibility

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	internal "gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

// Test

var _ CiVisibilityTest = (*tslvTest)(nil)

type tslvTest struct {
	ciVisibilityCommon
	suite *tslvTestSuite
	name  string
}

func createTest(suite *tslvTestSuite, name string, startTime time.Time) CiVisibilityTest {
	if suite == nil {
		return nil
	}

	operationName := "test"
	if suite.module.framework != "" {
		operationName = fmt.Sprintf("%s.%s", strings.ToLower(suite.module.framework), operationName)
	}

	resourceName := fmt.Sprintf("%s.%s", suite.name, name)

	// test tags should include also the suite, module and session tags so the backend can calculate the suite, module and session fingerprint from the test
	testTags := append(suite.tags, tracer.Tag(constants.TestName, name))
	testOpts := append(fillCommonTags([]tracer.StartSpanOption{
		tracer.ResourceName(resourceName),
		tracer.SpanType(constants.SpanTypeTest),
		tracer.StartTime(startTime),
	}), testTags...)

	span, ctx := tracer.StartSpanFromContext(context.Background(), operationName, testOpts...)
	if suite.module.session != nil {
		span.SetTag(constants.TestSessionIdTagName, fmt.Sprint(suite.module.session.sessionId))
	}
	span.SetTag(constants.TestModuleIdTagName, fmt.Sprint(suite.module.moduleId))
	span.SetTag(constants.TestSuiteIdTagName, fmt.Sprint(suite.suiteId))

	t := &tslvTest{
		suite: suite,
		name:  name,
		ciVisibilityCommon: ciVisibilityCommon{
			startTime: startTime,
			tags:      testTags,
			span:      span,
			ctx:       ctx,
		},
	}

	// We need to ensure to close everything before ci visibility is exiting.
	// In ci visibility mode we try to never lose data
	internal.PushCiVisibilityCloseAction(func() { t.Close(StatusFail) })

	return t
}

func (t *tslvTest) Name() string                  { return t.name }
func (t *tslvTest) Suite() CiVisibilityTestSuite  { return t.suite }
func (t *tslvTest) Close(status TestResultStatus) { t.CloseWithFinishTime(status, time.Now()) }
func (t *tslvTest) CloseWithFinishTime(status TestResultStatus, finishTime time.Time) {
	t.CloseWithFinishTimeAndSkipReason(status, finishTime, "")
}
func (t *tslvTest) CloseWithFinishTimeAndSkipReason(status TestResultStatus, finishTime time.Time, skipReason string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return
	}

	switch status {
	case StatusPass:
		t.span.SetTag(constants.TestStatus, constants.TestStatusPass)
	case StatusFail:
		t.span.SetTag(constants.TestStatus, constants.TestStatusFail)
	case StatusSkip:
		t.span.SetTag(constants.TestStatus, constants.TestStatusSkip)
	}

	if skipReason != "" {
		t.span.SetTag(constants.TestSkipReason, skipReason)
	}

	t.span.Finish(tracer.FinishTime(finishTime))
	t.closed = true
}
func (t *tslvTest) SetTestFunc(fn *runtime.Func) {
	if fn == nil {
		return
	}

	file, line := fn.FileLine(fn.Entry())
	file = utils.GetRelativePathFromCiTagsSourceRoot(file)
	t.SetTag(constants.TestSourceFile, file)
	t.SetTag(constants.TestSourceStartLine, line)

	codeOwners := utils.GetCodeOwners()
	if codeOwners != nil {
		match := codeOwners.Match("/" + file)
		if match != nil {
			t.SetTag(constants.TestCodeOwners, match.GetOwnersString())
		}
	}
}
