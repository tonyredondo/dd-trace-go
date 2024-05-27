// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"gopkg.in/DataDog/dd-trace-go.v1/internal/civisibility/constants"
	logger "gopkg.in/DataDog/dd-trace-go.v1/internal/log"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/osinfo"
)

var (
	// ciTags holds the CI/CD environment variable information.
	ciTags      map[string]string
	ciTagsMutex sync.Mutex

	// codeowners holds the parsed CODEOWNERS file data.
	codeowners      *CodeOwners
	codeownersMutex sync.Mutex
)

// GetCiTags retrieves and caches the CI/CD tags from environment variables.
// It initializes the ciTags map if it is not already initialized.
// This function is thread-safe due to the use of a mutex.
//
// Returns:
//
//	A map[string]string containing the CI/CD tags.
func GetCiTags() map[string]string {
	ciTagsMutex.Lock()
	defer ciTagsMutex.Unlock()

	if ciTags == nil {
		ciTags = createCiTagsMap()
	}

	return ciTags
}

// GetRelativePathFromCiTagsSourceRoot calculates the relative path from the CI workspace root to the specified path.
// If the CI workspace root is not available in the tags, it returns the original path.
//
// Parameters:
//
//	path - The absolute or relative file path for which the relative path should be calculated.
//
// Returns:
//
//	The relative path from the CI workspace root to the specified path, or the original path if an error occurs.
func GetRelativePathFromCiTagsSourceRoot(path string) string {
	tags := GetCiTags()
	if v, ok := tags[constants.CIWorkspacePath]; ok {
		relPath, err := filepath.Rel(v, path)
		if err == nil {
			return relPath
		}
	}

	return path
}

// GetCodeOwners retrieves and caches the CODEOWNERS data.
// It looks for the CODEOWNERS file in various standard locations within the CI workspace.
// This function is thread-safe due to the use of a mutex.
//
// Returns:
//
//	A pointer to a CodeOwners struct containing the parsed CODEOWNERS data, or nil if not found.
func GetCodeOwners() *CodeOwners {
	codeownersMutex.Lock()
	defer codeownersMutex.Unlock()

	if codeowners != nil {
		return codeowners
	}

	tags := GetCiTags()
	if v, ok := tags[constants.CIWorkspacePath]; ok {
		paths := []string{
			filepath.Join(v, "CODEOWNERS"),
			filepath.Join(v, ".github", "CODEOWNERS"),
			filepath.Join(v, ".gitlab", "CODEOWNERS"),
			filepath.Join(v, ".docs", "CODEOWNERS"),
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				codeowners, err = NewCodeOwners(path)
				if err != nil {
					logger.Debug("Error parsing codeowners: %s", err)
				}
			}
		}
	}

	return nil
}

// createCiTagsMap creates a map of CI/CD tags by extracting information from environment variables and the local Git repository.
// It also adds OS and runtime information to the tags.
//
// Returns:
//
//	A map[string]string containing the extracted CI/CD tags.
func createCiTagsMap() map[string]string {
	localTags := getProviderTags()
	localTags[constants.OSPlatform] = runtime.GOOS
	localTags[constants.OSVersion] = osinfo.OSVersion()
	localTags[constants.OSArchitecture] = runtime.GOARCH
	localTags[constants.RuntimeName] = runtime.Compiler
	localTags[constants.RuntimeVersion] = runtime.Version()

	gitData, _ := getLocalGitData()

	// Populate Git metadata from the local Git repository if not already present in localTags
	if _, ok := localTags[constants.CIWorkspacePath]; !ok {
		localTags[constants.CIWorkspacePath] = gitData.SourceRoot
	}
	if _, ok := localTags[constants.GitRepositoryURL]; !ok {
		localTags[constants.GitRepositoryURL] = gitData.RepositoryUrl
	}
	if _, ok := localTags[constants.GitCommitSHA]; !ok {
		localTags[constants.GitCommitSHA] = gitData.CommitSha
	}
	if _, ok := localTags[constants.GitBranch]; !ok {
		localTags[constants.GitBranch] = gitData.Branch
	}

	// If the commit SHA matches, populate additional Git metadata
	if localTags[constants.GitCommitSHA] == gitData.CommitSha {
		if _, ok := localTags[constants.GitCommitAuthorDate]; !ok {
			localTags[constants.GitCommitAuthorDate] = gitData.AuthorDate.String()
		}
		if _, ok := localTags[constants.GitCommitAuthorName]; !ok {
			localTags[constants.GitCommitAuthorName] = gitData.AuthorName
		}
		if _, ok := localTags[constants.GitCommitAuthorEmail]; !ok {
			localTags[constants.GitCommitAuthorEmail] = gitData.AuthorEmail
		}
		if _, ok := localTags[constants.GitCommitCommitterDate]; !ok {
			localTags[constants.GitCommitCommitterDate] = gitData.CommitterDate.String()
		}
		if _, ok := localTags[constants.GitCommitCommitterName]; !ok {
			localTags[constants.GitCommitCommitterName] = gitData.CommitterName
		}
		if _, ok := localTags[constants.GitCommitCommitterEmail]; !ok {
			localTags[constants.GitCommitCommitterEmail] = gitData.CommitterEmail
		}
		if _, ok := localTags[constants.GitCommitMessage]; !ok {
			localTags[constants.GitCommitMessage] = gitData.CommitMessage
		}
	}

	return localTags
}
