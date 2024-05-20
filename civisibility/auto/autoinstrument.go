// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package auto

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"gopkg.in/DataDog/dd-trace-go.v1/civisibility"
)

var (
	contextMutex sync.RWMutex
	contextMap   = map[*testing.T]context.Context{}
)

// Implementation for auto instrumentation

func Run(t *testing.T, name string, f func(t *testing.T)) bool {
	return t.Run(name, func(t *testing.T) {
		_, finish := civisibility.StartTestWithContext(GetContext(t), t, civisibility.WithOriginalTestFunc(f))
		defer finish()
		f(t)
	})
}

func RunM(m *testing.M) int {

	// Let's access to the inner Test array and instrument them
	internalTests := getInternalTestArray(m)
	if internalTests != nil {
		newTestArray := make([]testing.InternalTest, len(*internalTests))
		for idx, test := range *internalTests {
			testFn := test.F
			newTestArray[idx] = testing.InternalTest{
				Name: test.Name,
				F: func(t *testing.T) {
					_, finish := civisibility.StartTestWithContext(GetContext(t), t, civisibility.WithOriginalTestFunc(testFn))
					defer finish()
					testFn(t)
				},
			}
		}
		*internalTests = newTestArray
	}

	return civisibility.Run(m)
}

func RunTestMain(m *testing.M) {
	os.Exit(RunM(m))
}

func GetContext(t *testing.T) context.Context {
	// Read lock
	contextMutex.RLock()
	if ctx, ok := contextMap[t]; ok {
		return ctx
	}
	contextMutex.RUnlock()

	// Write lock
	ctx := context.Background()
	contextMutex.Lock()
	contextMap[t] = ctx
	contextMutex.Unlock()
	return ctx
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
