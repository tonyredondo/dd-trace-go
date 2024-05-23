// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package testing

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"

	"gopkg.in/DataDog/dd-trace-go.v1/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	internal "gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

const (
	testFramework = "golang.org/pkg/testing"
)

var (
	session civisibility.CiVisibilityTestSession

	testInfos       []*testingTInfo
	modulesCounters = map[string]*int32{}
	suitesCounters  = map[string]*int32{}
)

type testingTInfo struct {
	session      civisibility.CiVisibilityTestSession
	originalFunc func(*testing.T)
	moduleName   string
	suiteName    string
	testName     string
}

type M struct {
	*testing.M
}

func (m *M) Run() int {
	internal.EnsureCiVisibilityInitialization()
	defer internal.ExitCiVisibility()

	session = civisibility.CreateTestSession()

	// Let's access to the inner Test array and instrument them
	internalTests := getInternalTestArray(m.M)
	if internalTests != nil {

		// Extract info from internal tests
		testInfos = make([]*testingTInfo, len(*internalTests))
		for idx, test := range *internalTests {
			moduleName, suiteName := utils.GetModuleAndSuiteName(reflect.Indirect(reflect.ValueOf(test.F)).Pointer())
			testInfo := &testingTInfo{
				session:      session,
				originalFunc: test.F,
				moduleName:   moduleName,
				suiteName:    suiteName,
				testName:     test.Name,
			}

			if _, ok := modulesCounters[moduleName]; !ok {
				var v int32 = 0
				modulesCounters[moduleName] = &v
			}
			atomic.AddInt32(modulesCounters[moduleName], 1)

			if _, ok := suitesCounters[suiteName]; !ok {
				var v int32 = 0
				suitesCounters[suiteName] = &v
			}
			atomic.AddInt32(suitesCounters[suiteName], 1)

			testInfos[idx] = testInfo
		}

		// Create new instrumented internal tests
		newTestArray := make([]testing.InternalTest, len(*internalTests))
		for idx, testInfo := range testInfos {
			newTestArray[idx] = testing.InternalTest{
				Name: testInfo.testName,
				F:    executeInternalTest(testInfo),
			}
		}
		*internalTests = newTestArray
	}

	var exitCode = m.M.Run()

	session.Close(exitCode)
	return exitCode
}

func executeInternalTest(testInfo *testingTInfo) func(*testing.T) {
	originalFunc := runtime.FuncForPC(reflect.Indirect(reflect.ValueOf(testInfo.originalFunc)).Pointer())
	return func(t *testing.T) {
		module := session.GetOrCreateModuleWithFramework(testInfo.moduleName, testFramework, runtime.Version())
		suite := module.GetOrCreateSuite(testInfo.suiteName)
		test := suite.CreateTest(testInfo.testName)
		test.SetTestFunc(originalFunc)
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
				test.SetTag(ext.Error, t.Failed())
				if t.Failed() {
					test.Close(civisibility.StatusFail)
				} else if t.Skipped() {
					test.Close(civisibility.StatusSkip)
				} else {
					test.Close(civisibility.StatusPass)
				}

				checkModuleAndSuite(module, suite)
			}
		}()

		testInfo.originalFunc(t)
	}
}

func RunM(m *testing.M) int {
	ddM := M{M: m}
	return ddM.Run()
}

func RunAndExit(m *testing.M) {
	os.Exit(RunM(m))
}

type T struct {
	*testing.T
	test civisibility.CiVisibilityTest
}

func (ddt *T) Run(name string, f func(*testing.T)) bool {
	fReflect := reflect.Indirect(reflect.ValueOf(f))
	moduleName, suiteName := utils.GetModuleAndSuiteName(fReflect.Pointer())
	originalFunc := runtime.FuncForPC(fReflect.Pointer())
	// let's increment the test count in the module
	atomic.AddInt32(modulesCounters[moduleName], 1)
	// let's increment the test count in the suite
	atomic.AddInt32(suitesCounters[suiteName], 1)

	return ddt.T.Run(name, func(t *testing.T) {
		module := session.GetOrCreateModuleWithFramework(moduleName, testFramework, runtime.Version())
		suite := module.GetOrCreateSuite(suiteName)
		ddt.test = suite.CreateTest(t.Name())
		ddt.test.SetTestFunc(originalFunc)
		defer func() {
			if r := recover(); r != nil {
				// Panic handling
				ddt.test.SetErrorInfo("panic", fmt.Sprint(r), utils.GetStacktrace(2))
				ddt.test.Close(civisibility.StatusFail)
				checkModuleAndSuite(module, suite)
				internal.ExitCiVisibility()
				panic(r)
			} else {
				// Normal finalization
				ddt.test.SetTag(ext.Error, t.Failed())
				if t.Failed() {
					ddt.test.Close(civisibility.StatusFail)
				} else if t.Skipped() {
					ddt.test.Close(civisibility.StatusSkip)
				} else {
					ddt.test.Close(civisibility.StatusPass)
				}
				checkModuleAndSuite(module, suite)
			}
		}()

		f(t)
	})
}

func (ddt *T) Context() context.Context {
	if ddt.test != nil {
		return ddt.test.Context()
	}

	return context.Background()
}

func checkModuleAndSuite(module civisibility.CiVisibilityTestModule, suite civisibility.CiVisibilityTestSuite) {
	// If all tests in a suite has been executed we can close the suite
	if atomic.AddInt32(suitesCounters[suite.Name()], -1) <= 0 {
		suite.Close()
	}

	// If all tests in a module has been executed we can close the module
	if atomic.AddInt32(modulesCounters[module.Name()], -1) <= 0 {
		module.Close()
	}
}

// get the pointer to the internal test array
func getInternalTestArray(m *testing.M) *[]testing.InternalTest {
	indirectValue := reflect.Indirect(reflect.ValueOf(m))
	member := indirectValue.FieldByName("tests")
	if member.IsValid() {
		return (*[]testing.InternalTest)(unsafe.Pointer(member.UnsafeAddr()))
	}
	return nil
}
