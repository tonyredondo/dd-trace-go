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

// civisibilityCloseAction defines an action to be executed when CI visibility is closing.
type civisibilityCloseAction func()

var (
	// ciVisibilityInitializationOnce ensures we initialize the CI visibility tracer only once.
	ciVisibilityInitializationOnce sync.Once

	// closeActions holds CI visibility close actions.
	closeActions []civisibilityCloseAction

	// closeActionsMutex synchronizes access to closeActions.
	closeActionsMutex sync.Mutex
)

// EnsureCiVisibilityInitialization initializes the CI visibility tracer if it hasn't been initialized already.
func EnsureCiVisibilityInitialization() {
	ciVisibilityInitializationOnce.Do(func() {

		// Since calling this method indicates we are in CI Visibility mode, set the environment variable.
		_ = os.Setenv(constants.CiVisibilityEnabledEnvironmnetVariable, "1")

		// Preload all CI, Git, and CodeOwners tags.
		ciTags := utils.GetCiTags()
		_ = utils.GetCodeOwners()

		// Check if DD_SERVICE has been set; otherwise default to the repo name.
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

		// Initialize the tracer.
		tracer.Start(opts...)

		// Handle SIGINT and SIGTERM signals.
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signals
			ExitCiVisibility()
			os.Exit(1)
		}()
	})
}

// PushCiVisibilityCloseAction adds a close action to be executed when CI visibility exits.
func PushCiVisibilityCloseAction(action civisibilityCloseAction) {
	closeActionsMutex.Lock()
	defer closeActionsMutex.Unlock()
	closeActions = append([]civisibilityCloseAction{action}, closeActions...)
}

// ExitCiVisibility executes all registered close actions and stops the tracer.
func ExitCiVisibility() {
	closeActionsMutex.Lock()
	defer closeActionsMutex.Unlock()
	defer func() {
		closeActions = []civisibilityCloseAction{}

		tracer.Flush()
		tracer.Stop()
	}()
	for _, v := range closeActions {
		v()
	}
}
