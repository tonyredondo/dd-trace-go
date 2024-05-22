// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package civisibility

import (
	"context"
	"fmt"
	internal "gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
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

	modules map[string]CiVisibilityTestModule
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
	internal.EnsureCiVisibilityInitialization()

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

	s := &tslvTestSession{
		sessionId:        sessionId,
		command:          command,
		workingDirectory: workingDirectory,
		framework:        framework,
		modules:          map[string]CiVisibilityTestModule{},
		ciVisibilityCommon: ciVisibilityCommon{
			startTime: startTime,
			tags:      sessionTags,
			span:      span,
			ctx:       ctx,
		},
	}

	// We need to ensure to close everything before ci visibility is exiting.
	// In ci visibility mode we try to never lose data
	internal.PushCiVisibilityCloseAction(func() { s.Close(StatusFail) })

	return s
}

func (t *tslvTestSession) Command() string          { return t.command }
func (t *tslvTestSession) Framework() string        { return t.framework }
func (t *tslvTestSession) WorkingDirectory() string { return t.workingDirectory }
func (t *tslvTestSession) Close(exitCode int)       { t.CloseWithFinishTime(exitCode, time.Now()) }
func (t *tslvTestSession) CloseWithFinishTime(exitCode int, finishTime time.Time) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return
	}

	for _, m := range t.modules {
		m.Close()
	}
	t.modules = map[string]CiVisibilityTestModule{}

	t.span.SetTag(constants.TestCommandExitCode, exitCode)
	if exitCode == 0 {
		t.span.SetTag(constants.TestStatus, constants.TestStatusPass)
	} else {
		t.span.SetTag(constants.TestStatus, constants.TestStatusFail)
	}

	t.span.Finish(tracer.FinishTime(finishTime))
	t.closed = true

	tracer.Flush()
}
func (t *tslvTestSession) GetOrCreateModule(name string) CiVisibilityTestModule {
	return t.GetOrCreateModuleWithFramework(name, "", "")
}
func (t *tslvTestSession) GetOrCreateModuleWithFramework(name string, framework string, frameworkVersion string) CiVisibilityTestModule {
	return t.GetOrCreateModuleWithFrameworkAndStartTime(name, framework, frameworkVersion, time.Now())
}
func (t *tslvTestSession) GetOrCreateModuleWithFrameworkAndStartTime(name string, framework string, frameworkVersion string, startTime time.Time) CiVisibilityTestModule {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var mod CiVisibilityTestModule
	if v, ok := t.modules[name]; ok {
		mod = v
	} else {
		mod = createTestModule(t, name, framework, frameworkVersion, startTime)
		t.modules[name] = mod
	}

	return mod
}
