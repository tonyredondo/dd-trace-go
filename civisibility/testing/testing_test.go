// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package testing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	ddhttp "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func TestMain(m *testing.M) {
	RunAndExit(m)
}

func TestMyTest01(t *testing.T) {
	t.Log("My First Test")
}

func TestMyTest02(ot *testing.T) {
	ot.Log("My First Test 2")
	t := GetTest(ot)

	t.Run("sub01", func(oT2 *testing.T) {

		t2 := GetTest(oT2)

		t2.Log("From sub01")
		t2.Run("sub03", func(t3 *testing.T) {
			t3.Log("From sub03")
		})
	})
}

func Test_Foo(t *testing.T) {
	ddt := GetTest(t)

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
		ddt.Run(test.name, func(t *testing.T) {
			t.Log(test.name)
		})
	}
}

func TestWithExternalCalls(oT *testing.T) {
	t := GetTest(oT)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}))
	defer s.Close()

	t.Run("default", func(t *testing.T) {
		ctx := GetTest(t).Context()

		rt := ddhttp.WrapRoundTripper(http.DefaultTransport)
		client := &http.Client{
			Transport: rt,
		}

		req, err := http.NewRequest("GET", s.URL+"/hello/world", nil)
		if err != nil {
			t.FailNow()
		}

		req = req.WithContext(ctx)

		client.Do(req)
	})

	t.Run("custom-name", func(t *testing.T) {
		ctx := GetTest(t).Context()
		span, _ := ddtracer.SpanFromContext(ctx)

		customNamer := func(req *http.Request) string {
			value := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
			span.SetTag("customNamer.Value", value)
			return value
		}

		rt := ddhttp.WrapRoundTripper(http.DefaultTransport, ddhttp.RTWithResourceNamer(customNamer))
		client := &http.Client{
			Transport: rt,
		}

		req, err := http.NewRequest("GET", s.URL+"/hello/world", nil)
		if err != nil {
			t.FailNow()
		}

		req = req.WithContext(ctx)

		client.Do(req)
	})
}

func TestSkip(t *testing.T) {
	GetTest(t).Skip("Nothing to do here, skipping!")
}

func TestFail(t *testing.T) {
	GetTest(t).Fail()
}

func TestError(t *testing.T) {
	GetTest(t).Error("This is my: ", "Error")
}
