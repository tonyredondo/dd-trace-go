// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

//go:build !go1.20

package gotesting

import "testing"

// getCoverage prior to go1.20 the old coverage format is the default so we can use the normal testing.Coverage call
func getCoverage() float64 {
	return testing.Coverage()
}
