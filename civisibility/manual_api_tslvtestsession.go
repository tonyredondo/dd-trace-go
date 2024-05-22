// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package civisibility

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

// Test Session

var _ CiVisibilityTestSession = (*tslvTestSession)(nil)

type tslvTestSession struct {
	ciVisibilityCommon
	sessionId        uint64
	command          string
	workingDirectory string
	framework        string
}

func CreateTestSession() CiVisibilityTestSession {
	cmd := strings.Join(os.Args, " ")
	wd, err := os.Getwd()
	if err == nil {
		wd = utils.GetRelativePathFromCiTagsSourceRoot(wd)
	}
	return CreateTestSessionWith(cmd, wd, "", time.Now())
}
func CreateTestSessionWith(command string, workingDirectory string, framework string, startTime time.Time) CiVisibilityTestSession {
	// Let's ensure we have the ci visibility properly configured
	ensureCiVisibilityInitialization()

	operationName := "test_session"
	if framework != "" {
		operationName = fmt.Sprintf("%s.%s", strings.ToLower(framework), operationName)
	}

	resourceName := fmt.Sprintf("%s.%s", operationName, command)

	sessionTags := []tracer.StartSpanOption{
		tracer.Tag(constants.TestType, constants.TestTypeTest),
		tracer.Tag(constants.TestCommand, command),
		tracer.Tag(constants.TestCommandWorkingDirectory, workingDirectory),
	}

	testOpts := append(fillCommonTags([]tracer.StartSpanOption{
		tracer.ResourceName(resourceName),
		tracer.SpanType(constants.SpanTypeTestSession),
		tracer.StartTime(startTime),
	}), sessionTags...)

	span, ctx := tracer.StartSpanFromContext(context.Background(), operationName, testOpts...)
	sessionId := span.Context().SpanID()
	span.SetTag(constants.TestSessionIdTagName, fmt.Sprint(sessionId))

	session := &tslvTestSession{
		sessionId:        sessionId,
		command:          command,
		workingDirectory: workingDirectory,
		framework:        framework,
		ciVisibilityCommon: ciVisibilityCommon{
			startTime: startTime,
			tags:      sessionTags,
			span:      span,
			ctx:       ctx,
		},
	}

	// We need to ensure to close everything before ci visibility is exiting.
	// In ci visibility mode we try to never lose data
	pushCiVisibilityCloseAction(func() { session.Close(StatusFail) })

	return session
}

func (t *tslvTestSession) Command() string               { return t.command }
func (t *tslvTestSession) Framework() string             { return t.framework }
func (t *tslvTestSession) WorkingDirectory() string      { return t.workingDirectory }
func (t *tslvTestSession) Close(status TestResultStatus) { t.CloseWithFinishTime(status, time.Now()) }
func (t *tslvTestSession) CloseWithFinishTime(status TestResultStatus, finishTime time.Time) {
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

	t.span.Finish(tracer.FinishTime(finishTime))
	t.closed = true

	tracer.Flush()
}
func (t *tslvTestSession) CreateModule(name string) CiVisibilityTestModule {
	return t.CreateModuleWithFramework(name, "", "")
}
func (t *tslvTestSession) CreateModuleWithFramework(name string, framework string, frameworkVersion string) CiVisibilityTestModule {
	return t.CreateModuleWithFrameworkAndStartTime(name, framework, frameworkVersion, time.Now())
}
func (t *tslvTestSession) CreateModuleWithFrameworkAndStartTime(name string, framework string, frameworkVersion string, startTime time.Time) CiVisibilityTestModule {
	return createTestModule(t, name, framework, frameworkVersion, startTime)
}
