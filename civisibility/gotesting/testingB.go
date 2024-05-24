// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package gotesting

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
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
	ciVisibilityBenchmarks      = map[*testing.B]civisibility.CiVisibilityTest{}
	ciVisibilityBenchmarksMutex sync.RWMutex

	subBenchmarkAutoName      = "*--*AUTO*--*"
	subBenchmarkAutoNameRegex = regexp.MustCompile(`(?si)/\*--\*AUTO\*--\*.*`)
)

type B testing.B

func GetBenchmark(t *testing.B) *B { return (*B)(t) }

func (ddb *B) Run(name string, f func(*testing.B)) bool {
	fReflect := reflect.Indirect(reflect.ValueOf(f))
	moduleName, suiteName := utils.GetModuleAndSuiteName(fReflect.Pointer())
	originalFunc := runtime.FuncForPC(fReflect.Pointer())
	// let's increment the test count in the module
	atomic.AddInt32(modulesCounters[moduleName], 1)
	// let's increment the test count in the suite
	atomic.AddInt32(suitesCounters[suiteName], 1)

	pb := (*testing.B)(ddb)
	return pb.Run(subBenchmarkAutoName, func(b *testing.B) {
		// decrement level
		bpf := getBenchmarkPrivateFields(b)
		bpf.AddLevel(-1)

		startTime := time.Now()
		module := session.GetOrCreateModuleWithFrameworkAndStartTime(moduleName, testFramework, runtime.Version(), startTime)
		suite := module.GetOrCreateSuiteWithStartTime(suiteName, startTime)
		test := suite.CreateTestWithStartTime(fmt.Sprintf("%s/%s", pb.Name(), name), startTime)
		test.SetTestFunc(originalFunc)

		// restore the original name without the sub Benchmark auto name
		*bpf.name = subBenchmarkAutoNameRegex.ReplaceAllString(*bpf.name, "")

		// run original benchmark
		var iPfOfB *benchmarkPrivateFields
		var recoverFunc *func(r any)
		b.Run(name, func(b *testing.B) {
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
			*iPfOfB.benchFunc = f
			// set b to the civisibility test
			setCiVisibilityBenchmark(b, test)

			// enable the timer again
			b.StartTimer()
			// warmup the original func
			f(b)
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
			test.Close(civisibility.StatusFail)
			checkModuleAndSuite(module, suite)
			internal.ExitCiVisibility()
		}
		recoverFunc = &panicFunc

		// Normal finalization
		if iPfOfB.B.Failed() {
			test.SetTag(ext.Error, true)
			suite.SetTag(ext.Error, true)
			module.SetTag(ext.Error, true)
			test.CloseWithFinishTime(civisibility.StatusFail, endTime)
		} else if iPfOfB.B.Skipped() {
			test.CloseWithFinishTime(civisibility.StatusSkip, endTime)
		} else {
			test.CloseWithFinishTime(civisibility.StatusPass, endTime)
		}

		checkModuleAndSuite(module, suite)
	})
}

func (ddb *B) Context() context.Context {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		return ciTest.Context()
	}

	return context.Background()
}

func (ddb *B) Fail() {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fail", "failed test", utils.GetStacktrace(2))
	}

	b.Fail()
}

func (ddb *B) FailNow() {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.SetErrorInfo("FailNow", "failed test", utils.GetStacktrace(2))
	}

	internal.ExitCiVisibility()
	b.FailNow()
}

func (ddb *B) Error(args ...any) {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.SetErrorInfo("Error", fmt.Sprint(args...), utils.GetStacktrace(2))
	}

	b.Error(args...)
}

func (ddb *B) Errorf(format string, args ...any) {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.SetErrorInfo("Errorf", fmt.Sprintf(format, args...), utils.GetStacktrace(2))
	}

	b.Errorf(format, args...)
}

func (ddb *B) Fatal(args ...any) {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fatal", fmt.Sprint(args...), utils.GetStacktrace(2))
	}

	b.Fatal(args...)
}

func (ddb *B) Fatalf(format string, args ...any) {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.SetErrorInfo("Fatalf", fmt.Sprintf(format, args...), utils.GetStacktrace(2))
	}

	b.Fatalf(format, args...)
}

func (ddb *B) Skip(args ...any) {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.CloseWithFinishTimeAndSkipReason(civisibility.StatusSkip, time.Now(), fmt.Sprint(args...))
	}

	b.Skip(args...)
}

func (ddb *B) Skipf(format string, args ...any) {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.CloseWithFinishTimeAndSkipReason(civisibility.StatusSkip, time.Now(), fmt.Sprintf(format, args...))
	}

	b.Skipf(format, args...)
}

func (ddb *B) SkipNow() {
	b := (*testing.B)(ddb)
	ciTest := getCiVisibilityBenchmark(b)
	if ciTest != nil {
		ciTest.Close(civisibility.StatusSkip)
	}

	b.SkipNow()
}

func getCiVisibilityBenchmark(b *testing.B) civisibility.CiVisibilityTest {
	ciVisibilityBenchmarksMutex.RLock()
	defer ciVisibilityBenchmarksMutex.RUnlock()

	if v, ok := ciVisibilityBenchmarks[b]; ok {
		return v
	}

	return nil
}

func setCiVisibilityBenchmark(b *testing.B, ciTest civisibility.CiVisibilityTest) {
	ciVisibilityBenchmarksMutex.Lock()
	defer ciVisibilityBenchmarksMutex.Unlock()
	ciVisibilityBenchmarks[b] = ciTest
}
