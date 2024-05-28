// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package gotesting

import (
	"sync"
	"testing"
	"unsafe"
)

// Mock structs to simulate testing.M and testing.B
type mockTestingM struct {
	deps       testDeps
	tests      []testing.InternalTest
	benchmarks []testing.InternalBenchmark
}

// Dummy testDeps to emulate the memory layout of the original testing.M in order to get the right pointers to
// tests and benchmarks
type testDeps interface{}

func TestGetFieldPointerFrom(t *testing.T) {
	mockStruct := struct {
		privateField string
	}{
		privateField: "testValue",
	}

	ptr, err := getFieldPointerFrom(&mockStruct, "privateField")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ptr == nil {
		t.Fatal("Expected a valid pointer, got nil")
	}

	actualValue := (*string)(ptr)
	if *actualValue != mockStruct.privateField {
		t.Fatalf("Expected 'testValue', got %s", *actualValue)
	}

	*actualValue = "modified value"
	if *actualValue != mockStruct.privateField {
		t.Fatalf("Expected 'testValue', got %s", *actualValue)
	}

	_, err = getFieldPointerFrom(&mockStruct, "nonExistentField")
	if err == nil {
		t.Fatal("Expected an error for non-existent field, got nil")
	}
}

func TestGetInternalTestArray(t *testing.T) {
	mockM := &mockTestingM{
		tests: []testing.InternalTest{{Name: "Test1"}},
	}

	tests := getInternalTestArray((*testing.M)(unsafe.Pointer(mockM)))
	if tests == nil {
		t.Fatal("Expected a valid pointer to InternalTest array, got nil")
	}

	if len(*tests) != 1 || (*tests)[0].Name != "Test1" {
		t.Fatalf("Expected a single test named 'Test1', got %+v", *tests)
	}
}

func TestGetInternalBenchmarkArray(t *testing.T) {
	mockM := &mockTestingM{
		benchmarks: []testing.InternalBenchmark{{Name: "Benchmark1"}},
	}

	benchmarks := getInternalBenchmarkArray((*testing.M)(unsafe.Pointer(mockM)))
	if benchmarks == nil {
		t.Fatal("Expected a valid pointer to InternalBenchmark array, got nil")
	}

	if len(*benchmarks) != 1 || (*benchmarks)[0].Name != "Benchmark1" {
		t.Fatalf("Expected a single benchmark named 'Benchmark1', got %+v", *benchmarks)
	}
}

func TestCommonPrivateFields_AddLevel(t *testing.T) {
	mu := &sync.RWMutex{}
	level := 1
	commonFields := &commonPrivateFields{
		mu:    mu,
		level: &level,
	}

	newLevel := commonFields.AddLevel(1)
	if newLevel != 2 || newLevel != *commonFields.level {
		t.Fatalf("Expected level to be 2, got %d", newLevel)
	}

	newLevel = commonFields.AddLevel(-1)
	if newLevel != 1 || newLevel != *commonFields.level {
		t.Fatalf("Expected level to be 1, got %d", newLevel)
	}
}

func TestGetBenchmarkPrivateFields(t *testing.T) {
	b := &testing.B{}
	benchFields := getBenchmarkPrivateFields(b)
	if benchFields == nil {
		t.Fatal("Expected a valid benchmarkPrivateFields, got nil")
	}

	*benchFields.name = "BenchmarkTest"
	*benchFields.level = 1
	*benchFields.benchFunc = func(b *testing.B) {}
	*benchFields.result = testing.BenchmarkResult{}

	if benchFields.level == nil || *benchFields.level != 1 {
		t.Fatalf("Expected level to be 1, got %v", *benchFields.level)
	}

	if benchFields.name == nil || *benchFields.name != b.Name() {
		t.Fatalf("Expected name to be 'BenchmarkTest', got %v", *benchFields.name)
	}

	if benchFields.benchFunc == nil {
		t.Fatal("Expected benchFunc to be set, got nil")
	}

	if benchFields.result == nil {
		t.Fatal("Expected result to be set, got nil")
	}
}
