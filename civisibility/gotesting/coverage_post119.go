// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

//go:build go1.20

package gotesting

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"runtime/coverage"
	"strconv"
	"testing"
	_ "unsafe"

	logger "gopkg.in/DataDog/dd-trace-go.v1/internal/log"
)

//go:linkname runtime_coverage_processCoverTestDirInternal runtime/coverage.processCoverTestDirInternal
func runtime_coverage_processCoverTestDirInternal(dir string, cfile string, cm string, cpkg string, w io.Writer) error

// force the package to be included in the binary so the linker (in go:linkname) can find the symbols
var _ = coverage.ClearCounters

// getCoverage uses the internal `runtime/coverage.processCoverTestDirInternal` to process the coverage counters
// then parse the result and return the percentage value in float64
func getCoverage() float64 {
	goCoverDir := os.Getenv("GOCOVERDIR")
	if goCoverDir == "" {
		logger.Warn("GOCOVERDIR environment variable not set, coverage couldn't get extracted.")
		return testing.Coverage()
	}

	buffer := new(bytes.Buffer)
	err := runtime_coverage_processCoverTestDirInternal(goCoverDir, "", testing.CoverMode(), "", buffer)
	if err == nil {
		re := regexp.MustCompile(`(?si)coverage: (.*)%`)
		results := re.FindStringSubmatch(buffer.String())
		if len(results) == 2 {
			percentage, err := strconv.ParseFloat(results[1], 64)
			if err == nil {
				return percentage
			}
		}
	} else {
		logger.Warn("Error trying to get coverage stats: %s", err)
	}
	return testing.Coverage()
}
