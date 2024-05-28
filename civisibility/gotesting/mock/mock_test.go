// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package mock

import (
	"os"
	"testing"
	_ "unsafe"

	"gopkg.in/DataDog/dd-trace-go.v1/civisibility/gotesting"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility"
)

func TestMain(om *testing.M) {
	// Initialize civisibility using the mocktracer for testing
	tracer := civisibility.InitializeCiVisibilityMock()
	m := (*gotesting.M)(om)
	exitCode := m.Run()
	_ = tracer
	os.Exit(exitCode)
}

func TestSample1(t *testing.T) {

}
