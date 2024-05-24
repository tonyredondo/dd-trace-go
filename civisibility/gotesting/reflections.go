// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package gotesting

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func getFieldPointerFrom(value any, fieldName string) (unsafe.Pointer, error) {
	indirectValue := reflect.Indirect(reflect.ValueOf(value))
	member := indirectValue.FieldByName(fieldName)
	if member.IsValid() {
		return unsafe.Pointer(member.UnsafeAddr()), nil
	}

	return unsafe.Pointer(nil), errors.New("member is invalid")
}

// TESTING

// get the pointer to the internal test array
func getInternalTestArray(m *testing.M) *[]testing.InternalTest {
	if ptr, err := getFieldPointerFrom(m, "tests"); err == nil {
		return (*[]testing.InternalTest)(ptr)
	}
	return nil
}

// BENCHMARKS

// get the pointer to the internal benchmark array
func getInternalBenchmarkArray(m *testing.M) *[]testing.InternalBenchmark {
	if ptr, err := getFieldPointerFrom(m, "benchmarks"); err == nil {
		return (*[]testing.InternalBenchmark)(ptr)
	}
	return nil
}

type commonPrivateFields struct {
	mu       *sync.RWMutex
	level    *int
	start    *time.Time // Time test or benchmark started
	duration *time.Duration
	name     *string // Name of test or benchmark.
}

func (c *commonPrivateFields) AddLevel(delta int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	*c.level = *c.level + delta
	return *c.level
}

type benchmarkPrivateFields struct {
	commonPrivateFields
	B         *testing.B
	benchFunc *func(b *testing.B)
	result    *testing.BenchmarkResult
}

func getBenchmarkPrivateFields(b *testing.B) *benchmarkPrivateFields {
	benchFields := &benchmarkPrivateFields{
		B: b,
	}

	// common
	if ptr, err := getFieldPointerFrom(b, "mu"); err == nil {
		benchFields.mu = (*sync.RWMutex)(ptr)
	}
	if ptr, err := getFieldPointerFrom(b, "level"); err == nil {
		benchFields.level = (*int)(ptr)
	}
	if ptr, err := getFieldPointerFrom(b, "start"); err == nil {
		benchFields.start = (*time.Time)(ptr)
	}
	if ptr, err := getFieldPointerFrom(b, "duration"); err == nil {
		benchFields.duration = (*time.Duration)(ptr)
	}
	if ptr, err := getFieldPointerFrom(b, "name"); err == nil {
		benchFields.name = (*string)(ptr)
	}

	// benchmark
	if ptr, err := getFieldPointerFrom(b, "benchFunc"); err == nil {
		benchFields.benchFunc = (*func(b *testing.B))(ptr)
	}
	if ptr, err := getFieldPointerFrom(b, "result"); err == nil {
		benchFields.result = (*testing.BenchmarkResult)(ptr)
	}

	return benchFields
}
