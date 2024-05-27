// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package gotesting

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	internal "gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

const (
	testFramework = "golang.org/pkg/testing"
)

var (
	session civisibility.DdTestSession

	testInfos       []*testingTInfo
	benchmarkInfos  []*testingBInfo
	modulesCounters = map[string]*int32{}
	suitesCounters  = map[string]*int32{}
)

type (
	commonInfo struct {
		moduleName string
		suiteName  string
		testName   string
	}

	testingTInfo struct {
		commonInfo
		originalFunc func(*testing.T)
	}

	testingBInfo struct {
		commonInfo
		originalFunc func(b *testing.B)
	}

	M testing.M
)

func (ddm *M) Run() int {
	internal.EnsureCiVisibilityInitialization()
	defer internal.ExitCiVisibility()

	session = civisibility.CreateTestSession()

	m := (*testing.M)(ddm)

	// Access to the inner Test array and instrument them
	ddm.instrumentInternalTests(getInternalTestArray(m))

	// Access to the inner Benchmark array and instrument them
	ddm.instrumentInternalBenchmarks(getInternalBenchmarkArray(m))

	var exitCode = m.Run()
	coveragePercentage := getCoverage()
	if testing.CoverMode() != "" {
		session.SetTag(constants.CodeCoverageEnabledTagName, "true")
		session.SetTag(constants.CodeCoveragePercentageOfTotalLines, coveragePercentage)
	}

	session.Close(exitCode)
	return exitCode
}

func (ddm *M) instrumentInternalTests(internalTests *[]testing.InternalTest) {
	if internalTests != nil {
		// Extract info from internal tests
		testInfos = make([]*testingTInfo, len(*internalTests))
		for idx, test := range *internalTests {
			moduleName, suiteName := utils.GetModuleAndSuiteName(reflect.Indirect(reflect.ValueOf(test.F)).Pointer())
			testInfo := &testingTInfo{
				originalFunc: test.F,
				commonInfo: commonInfo{
					moduleName: moduleName,
					suiteName:  suiteName,
					testName:   test.Name,
				},
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
				F:    ddm.executeInternalTest(testInfo),
			}
		}
		*internalTests = newTestArray
	}
}

func (ddm *M) executeInternalTest(testInfo *testingTInfo) func(*testing.T) {
	originalFunc := runtime.FuncForPC(reflect.Indirect(reflect.ValueOf(testInfo.originalFunc)).Pointer())
	return func(t *testing.T) {
		module := session.GetOrCreateModuleWithFramework(testInfo.moduleName, testFramework, runtime.Version())
		suite := module.GetOrCreateSuite(testInfo.suiteName)
		test := suite.CreateTest(testInfo.testName)
		test.SetTestFunc(originalFunc)
		setCiVisibilityTest(t, test)
		defer func() {
			if r := recover(); r != nil {
				// Panic handling
				test.SetErrorInfo("panic", fmt.Sprint(r), utils.GetStacktrace(2))
				suite.SetTag(ext.Error, true)
				module.SetTag(ext.Error, true)
				test.Close(civisibility.ResultStatusFail)
				checkModuleAndSuite(module, suite)
				internal.ExitCiVisibility()
				panic(r)
			} else {
				// Normal finalization
				if t.Failed() {
					test.SetTag(ext.Error, true)
					suite.SetTag(ext.Error, true)
					module.SetTag(ext.Error, true)
					test.Close(civisibility.ResultStatusFail)
				} else if t.Skipped() {
					test.Close(civisibility.ResultStatusSkip)
				} else {
					test.Close(civisibility.ResultStatusPass)
				}

				checkModuleAndSuite(module, suite)
			}
		}()

		testInfo.originalFunc(t)
	}
}

func (ddm *M) instrumentInternalBenchmarks(internalBenchmarks *[]testing.InternalBenchmark) {
	if internalBenchmarks != nil {
		// Extract info from internal benchmarks
		benchmarkInfos = make([]*testingBInfo, len(*internalBenchmarks))
		for idx, benchmark := range *internalBenchmarks {
			moduleName, suiteName := utils.GetModuleAndSuiteName(reflect.Indirect(reflect.ValueOf(benchmark.F)).Pointer())
			benchmarkInfo := &testingBInfo{
				originalFunc: benchmark.F,
				commonInfo: commonInfo{
					moduleName: moduleName,
					suiteName:  suiteName,
					testName:   benchmark.Name,
				},
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

			benchmarkInfos[idx] = benchmarkInfo
		}

		// Create a new instrumented internal benchmarks
		newBenchmarkArray := make([]testing.InternalBenchmark, len(*internalBenchmarks))
		for idx, benchmarkInfo := range benchmarkInfos {
			newBenchmarkArray[idx] = testing.InternalBenchmark{
				Name: benchmarkInfo.testName,
				F:    ddm.executeInternalBenchmark(benchmarkInfo),
			}
		}

		*internalBenchmarks = newBenchmarkArray
	}
}

func (ddm *M) executeInternalBenchmark(benchmarkInfo *testingBInfo) func(*testing.B) {
	return func(b *testing.B) {

		// decrement level
		getBenchmarkPrivateFields(b).AddLevel(-1)

		startTime := time.Now()
		originalFunc := runtime.FuncForPC(reflect.Indirect(reflect.ValueOf(benchmarkInfo.originalFunc)).Pointer())
		module := session.GetOrCreateModuleWithFrameworkAndStartTime(benchmarkInfo.moduleName, testFramework, runtime.Version(), startTime)
		suite := module.GetOrCreateSuiteWithStartTime(benchmarkInfo.suiteName, startTime)
		test := suite.CreateTestWithStartTime(benchmarkInfo.testName, startTime)
		test.SetTestFunc(originalFunc)

		// run original benchmark
		var iPfOfB *benchmarkPrivateFields
		var recoverFunc *func(r any)
		b.Run(b.Name(), func(b *testing.B) {
			// stop the timer to do the initialization and replacements
			b.StopTimer()

			defer func() {
				if r := recover(); r != nil {
					if recoverFunc != nil {
						fn := *recoverFunc
						fn(r)
					}
					panic(r)
				}
			}()

			// enable allocations reporting
			b.ReportAllocs()
			// first time we get the private fields of the inner testing.B
			iPfOfB = getBenchmarkPrivateFields(b)
			// replace this function with the original one (this should be executed only once - the first iteration[b.run1])
			*iPfOfB.benchFunc = benchmarkInfo.originalFunc
			// set b to the civisibility test
			setCiVisibilityBenchmark(b, test)

			// enable the timer again
			b.StartTimer()
			// warmup the original func
			benchmarkInfo.originalFunc(b)
		})

		endTime := time.Now()
		results := iPfOfB.result
		test.SetBenchmarkData("duration", map[string]any{
			"run":  results.N,
			"mean": results.NsPerOp(),
		})
		test.SetBenchmarkData("memory_total_operations", map[string]any{
			"run":            results.N,
			"mean":           results.AllocsPerOp(),
			"statistics.max": results.MemAllocs,
		})
		test.SetBenchmarkData("mean_heap_allocations", map[string]any{
			"run":  results.N,
			"mean": results.AllocedBytesPerOp(),
		})
		test.SetBenchmarkData("total_heap_allocations", map[string]any{
			"run":  results.N,
			"mean": iPfOfB.result.MemBytes,
		})
		if len(results.Extra) > 0 {
			mapConverted := map[string]any{}
			for k, v := range results.Extra {
				mapConverted[k] = v
			}
			test.SetBenchmarkData("extra", mapConverted)
		}

		panicFunc := func(r any) {
			// Panic handling
			test.SetErrorInfo("panic", fmt.Sprint(r), utils.GetStacktrace(2))
			suite.SetTag(ext.Error, true)
			module.SetTag(ext.Error, true)
			test.Close(civisibility.ResultStatusFail)
			checkModuleAndSuite(module, suite)
			internal.ExitCiVisibility()
		}
		recoverFunc = &panicFunc

		// Normal finalization
		if iPfOfB.B.Failed() {
			test.SetTag(ext.Error, true)
			suite.SetTag(ext.Error, true)
			module.SetTag(ext.Error, true)
			test.CloseWithFinishTime(civisibility.ResultStatusFail, endTime)
		} else if iPfOfB.B.Skipped() {
			test.CloseWithFinishTime(civisibility.ResultStatusSkip, endTime)
		} else {
			test.CloseWithFinishTime(civisibility.ResultStatusPass, endTime)
		}

		checkModuleAndSuite(module, suite)
	}
}

func RunM(m *testing.M) int {
	return (*M)(m).Run()
}

func RunAndExit(m *testing.M) {
	os.Exit(RunM(m))
}

func checkModuleAndSuite(module civisibility.DdTestModule, suite civisibility.DdTestSuite) {
	// If all tests in a suite has been executed we can close the suite
	if atomic.AddInt32(suitesCounters[suite.Name()], -1) <= 0 {
		suite.Close()
	}

	// If all tests in a module has been executed we can close the module
	if atomic.AddInt32(modulesCounters[module.Name()], -1) <= 0 {
		module.Close()
	}
}
