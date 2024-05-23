// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package utils

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// GetModuleAndSuiteName gets the module name and suite name given a program counter.
// Uses runtime.FuncForPC internally to get the full func name of the program counter,
// and the file then it will split the string by the searching for the latest dot ('.')
// in the string that separate the full package name from the actual func name.
// Example 1:
//
//	input: github.com/DataDog/dd-sdk-go-testing.TestRun
//	output:
//	   module: github.com/DataDog/dd-sdk-go-testing
//	   suite: testing_test.go
//
// Example 2:
//
//	input: github.com/DataDog/dd-sdk-go-testing.TestRun.func1
//	output:
//	   module: github.com/DataDog/dd-sdk-go-testing
//	   suite: testing_test.go
func GetModuleAndSuiteName(pc uintptr) (module string, suite string) {
	funcValue := runtime.FuncForPC(pc)
	funcFullName := funcValue.Name()
	lastSlash := strings.LastIndexByte(funcFullName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	firstDot := strings.IndexByte(funcFullName[lastSlash:], '.') + lastSlash
	file, _ := funcValue.FileLine(funcValue.Entry())
	return funcFullName[:firstDot], filepath.Base(file)
}

func GetStacktrace(skip int) string {
	pcs := make([]uintptr, 256)
	total := runtime.Callers(skip+1, pcs)
	frames := runtime.CallersFrames(pcs[:total])
	buffer := new(bytes.Buffer)
	for {
		if frame, ok := frames.Next(); ok {
			_, _ = fmt.Fprintf(buffer, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		} else {
			break
		}

	}
	return buffer.String()
}
