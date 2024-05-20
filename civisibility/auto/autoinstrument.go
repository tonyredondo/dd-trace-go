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

type T testing.T
type M testing.M

// Implementation for auto instrumentation

func (t *T) Run(name string, f func(t *testing.T)) bool {
	return (*testing.T)(t).Run(name, func(innerT *testing.T) {
		_, finish := civisibility.StartTestWithContext(GetContext(innerT), innerT, civisibility.WithOriginalTestFunc(f))
		defer finish()
		f(innerT)
	})
}

func Run(t *testing.T, name string, f func(t *testing.T)) bool {
	return (*T)(t).Run(name, f)
}

func (m *M) Run() int {
	testingM := (*testing.M)(m)

	// Let's access to the inner Test array and instrument them
	internalTests := getInternalTestArray(testingM)
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

	return civisibility.Run(testingM)
}

func RunM(m *testing.M) int {
	return (*M)(m).Run()
}

func RunAndExit(m *testing.M) {
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
