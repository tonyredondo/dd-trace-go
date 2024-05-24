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

func GetTest(t *testing.T) *T {
	return (*T)(t)
}

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

func (ddt *T) Context() context.Context {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		return ciTest.Context()
	}

	return context.Background()
}

func (ddt *T) Fail() {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fail", "failed test", utils.GetStacktrace(2))
	}

	t.Fail()
}

func (ddt *T) FailNow() {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("FailNow", "failed test", utils.GetStacktrace(2))
	}

	internal.ExitCiVisibility()
	t.FailNow()
}

func (ddt *T) Error(args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Error", fmt.Sprint(args...), utils.GetStacktrace(2))
	}

	t.Error(args...)
}

func (ddt *T) Errorf(format string, args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Errorf", fmt.Sprintf(format, args...), utils.GetStacktrace(2))
	}

	t.Errorf(format, args...)
}

func (ddt *T) Fatal(args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fatal", fmt.Sprint(args...), utils.GetStacktrace(2))
	}

	t.Fatal(args...)
}

func (ddt *T) Fatalf(format string, args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fatalf", fmt.Sprintf(format, args...), utils.GetStacktrace(2))
	}

	t.Fatalf(format, args...)
}

func (ddt *T) Skip(args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.CloseWithFinishTimeAndSkipReason(civisibility.StatusSkip, time.Now(), fmt.Sprint(args...))
	}

	t.Skip(args...)
}

func (ddt *T) Skipf(format string, args ...any) {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.CloseWithFinishTimeAndSkipReason(civisibility.StatusSkip, time.Now(), fmt.Sprintf(format, args...))
	}

	t.Skipf(format, args...)
}

func (ddt *T) SkipNow() {
	t := (*testing.T)(ddt)
	ciTest := getCiVisibilityTest(t)
	if ciTest != nil {
		ciTest.Close(civisibility.StatusSkip)
	}

	t.SkipNow()
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
