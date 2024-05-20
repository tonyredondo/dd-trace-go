// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package auto

import (
	"testing"
)

func TestMain(m *testing.M) {
	RunAndExit(m)
}

func TestMyTest01(t *testing.T) {
	t.Log("My First Test")
}

func TestMyTest02(t *testing.T) {
	t.Log("My First Test 2")

	Run(t, "sub01", func(oT2 *testing.T) {

		t2 := T{oT2}

		t2.Log("From sub01")
		t2.Run("sub03", func(t3 *testing.T) {
			t3.Log("From sub03")
		})
	})
}
