// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package testing

import (
	"testing"
)

func TestMain(m *testing.M) {
	RunAndExit(m)
}

func TestMyTest01(t *testing.T) {
	t.Log("My First Test")
}

func TestMyTest02(ot *testing.T) {
	ot.Log("My First Test 2")
	t := T{T: ot}

	t.Run("sub01", func(oT2 *testing.T) {

		t2 := T{T: oT2}

		t2.Log("From sub01")
		t2.Run("sub03", func(t3 *testing.T) {
			t3.Log("From sub03")
		})
	})
}

func Test_Foo(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"yellow should return color", "yellow", "color"},
		{"banana should return fruit", "banana", "fruit"},
		{"duck should return animal", "duck", "animal"},
	}
	for _, test := range tests {
		ddt := T{T: t}
		ddt.Run(test.name, func(t *testing.T) {
			t.Log(test.name)
		})
	}
}
