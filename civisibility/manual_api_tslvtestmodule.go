// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package civisibility

import (
	"context"
	"fmt"
	internal "gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
	"strings"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
)

// Test Module

var _ CiVisibilityTestModule = (*tslvTestModule)(nil)

type tslvTestModule struct {
	ciVisibilityCommon
	session   *tslvTestSession
	moduleId  uint64
	name      string
	framework string

	suites map[string]CiVisibilityTestSuite
}

func createTestModule(session *tslvTestSession, name string, framework string, frameworkVersion string, startTime time.Time) CiVisibilityTestModule {
	// Let's ensure we have the ci visibility properly configured
	internal.EnsureCiVisibilityInitialization()

	operationName := "test_module"
	if framework != "" {
		operationName = fmt.Sprintf("%s.%s", strings.ToLower(framework), operationName)
	}

	resourceName := name

	var sessionTags []tracer.StartSpanOption
	if session != nil {
		sessionTags = session.tags
	}

	// module tags should include also the session tags so the backend can calculate the session fingerprint from the module
	moduleTags := append(sessionTags, []tracer.StartSpanOption{
		tracer.Tag(constants.TestType, constants.TestTypeTest),
		tracer.Tag(constants.TestModule, name),
		tracer.Tag(constants.TestFramework, framework),
		tracer.Tag(constants.TestFrameworkVersion, frameworkVersion),
	}...)

	testOpts := append(fillCommonTags([]tracer.StartSpanOption{
		tracer.ResourceName(resourceName),
		tracer.SpanType(constants.SpanTypeTestModule),
		tracer.StartTime(startTime),
	}), moduleTags...)

	span, ctx := tracer.StartSpanFromContext(context.Background(), operationName, testOpts...)
	moduleId := span.Context().SpanID()
	if session != nil {
		span.SetTag(constants.TestSessionIdTagName, fmt.Sprint(session.sessionId))
	}
	span.SetTag(constants.TestModuleIdTagName, fmt.Sprint(moduleId))

	module := &tslvTestModule{
		session:   session,
		moduleId:  moduleId,
		name:      name,
		framework: framework,
		suites:    map[string]CiVisibilityTestSuite{},
		ciVisibilityCommon: ciVisibilityCommon{
			startTime: startTime,
			tags:      moduleTags,
			span:      span,
			ctx:       ctx,
		},
	}

	// We need to ensure to close everything before ci visibility is exiting.
	// In ci visibility mode we try to never lose data
	internal.PushCiVisibilityCloseAction(func() { module.Close() })

	return module
}

func (t *tslvTestModule) Name() string                     { return t.name }
func (t *tslvTestModule) Framework() string                { return t.framework }
func (t *tslvTestModule) Session() CiVisibilityTestSession { return t.session }
func (t *tslvTestModule) Close()                           { t.CloseWithFinishTime(time.Now()) }
func (t *tslvTestModule) CloseWithFinishTime(finishTime time.Time) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return
	}

	for _, suite := range t.suites {
		suite.Close()
	}
	t.suites = map[string]CiVisibilityTestSuite{}

	t.span.Finish(tracer.FinishTime(finishTime))
	t.closed = true
}
func (t *tslvTestModule) GetOrCreateSuite(name string) CiVisibilityTestSuite {
	return t.GetOrCreateSuiteWithStartTime(name, time.Now())
}
func (t *tslvTestModule) GetOrCreateSuiteWithStartTime(name string, startTime time.Time) CiVisibilityTestSuite {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var suite CiVisibilityTestSuite
	if v, ok := t.suites[name]; ok {
		suite = v
	} else {
		suite = createTestSuite(t, name, startTime)
		t.suites[name] = suite
	}

	return suite
}
