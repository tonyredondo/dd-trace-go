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
	// tags contains information detected from CI/CD environment variables.
	ciTags      map[string]string
	ciTagsMutex sync.Mutex

	codeowners      *CodeOwners
	codeownersMutex sync.Mutex
)

func GetCiTags() map[string]string {
	ciTagsMutex.Lock()
	defer ciTagsMutex.Unlock()

	if ciTags == nil {
		ciTags = createCiTagsMap()
	}

	return ciTags
}

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

func createCiTagsMap() map[string]string {
	localTags := getProviderTags()
	localTags[constants.OSPlatform] = runtime.GOOS
	localTags[constants.OSVersion] = osinfo.OSVersion()
	localTags[constants.OSArchitecture] = runtime.GOARCH
	localTags[constants.RuntimeName] = runtime.Compiler
	localTags[constants.RuntimeVersion] = runtime.Version()

	gitData, _ := getLocalGitData()

	// Guess Git metadata from a local Git repository otherwise.
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
