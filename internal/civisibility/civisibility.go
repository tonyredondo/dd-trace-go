// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package civisibility

import (
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/utils"
)

type (
	// civisibilityCloseAction action to be executed when ci visibility is closing
	civisibilityCloseAction func()
)

var (
	// ciVisibilityInitializationOnce ensure we initialize the ci visibility tracer only once
	ciVisibilityInitializationOnce sync.Once

	// closeActions ci visibility close actions
	closeActions []civisibilityCloseAction

	// closeActionsMutex ci visibility close actions mutex
	closeActionsMutex sync.Mutex
)

func EnsureCiVisibilityInitialization() {
	ciVisibilityInitializationOnce.Do(func() {
		// Preload all CI and Git tags.
		ciTags := utils.GetCiTags()

		// Check if DD_SERVICE has been set; otherwise we default to repo name.
		var opts []tracer.StartOption
		if v := os.Getenv("DD_SERVICE"); v == "" {
			if repoUrl, ok := ciTags[constants.GitRepositoryURL]; ok {
				// regex to sanitize the repository url to be used as a service name
				repoRegex := regexp.MustCompile(`(?m)/([a-zA-Z0-9\\\-_.]*)$`)
				matches := repoRegex.FindStringSubmatch(repoUrl)
				if len(matches) > 1 {
					repoUrl = strings.TrimSuffix(matches[1], ".git")
				}
				opts = append(opts, tracer.WithService(repoUrl))
			}
		}

		// Initialize tracer
		tracer.Start(opts...)

		// Handle SIGINT and SIGTERM
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signals
			ExitCiVisibility()
			os.Exit(1)
		}()
	})
}

func PushCiVisibilityCloseAction(action civisibilityCloseAction) {
	closeActionsMutex.Lock()
	defer closeActionsMutex.Unlock()
	closeActions = append([]civisibilityCloseAction{action}, closeActions...)
}

func ExitCiVisibility() {
	closeActionsMutex.Lock()
	defer closeActionsMutex.Unlock()
	for _, v := range closeActions {
		v()
	}
	closeActions = []civisibilityCloseAction{}

	tracer.Flush()
	tracer.Stop()
}