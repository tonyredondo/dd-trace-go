// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package constants

const (
	// CiVisibilityEnabledEnvironmnetVariable indicates if CI Visibility mode is enabled
	CiVisibilityEnabledEnvironmnetVariable = "DD_CIVISIBILITY_ENABLED"

	// CiVisibilityAgentlessEnabledEnvironmentVariable indicate if CI Visibility agentless mode is enabled
	CiVisibilityAgentlessEnabledEnvironmentVariable = "DD_CIVISIBILITY_AGENTLESS_ENABLED"

	// CiVisibilityAgentlessUrlEnvironmentVariable forces the agentless url to a custom one
	CiVisibilityAgentlessUrlEnvironmentVariable = "DD_CIVISIBILITY_AGENTLESS_URL"

	// ApiKeyEnvironmentVariable indicates the Api key to be used for agentless intake
	ApiKeyEnvironmentVariable = "DD_API_KEY"
)
