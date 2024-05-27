// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	logger "gopkg.in/DataDog/dd-trace-go.v1/internal/log"
)

// CodeOwners represents a structured data type that holds sections of code owners.
// Each section maps to a slice of entries, where each entry includes a pattern and a list of owners.
type CodeOwners struct {
	Sections map[string][]Entry
}

// Entry represents a single entry in a CODEOWNERS file.
// It includes the pattern for matching files, the list of owners, and the section to which it belongs.
type Entry struct {
	Pattern string
	Owners  []string
	Section string
}

// NewCodeOwners creates a new instance of CodeOwners by parsing a CODEOWNERS file located at the given filePath.
// It returns an error if the file cannot be read or parsed properly.
func NewCodeOwners(filePath string) (*CodeOwners, error) {
	if filePath == "" {
		return nil, fmt.Errorf("filePath cannot be empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
		if err != nil && !errors.Is(os.ErrClosed, err) {
			logger.Warn("Error closing codeowners file: %s", err.Error())
		}
	}()

	var entriesList []Entry
	var sectionsList []string
	var currentSectionName string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Identify section headers, which are lines enclosed in square brackets
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSectionName = line[1 : len(line)-1]
			foundSectionName := findSectionIgnoreCase(sectionsList, currentSectionName)
			if foundSectionName == "" {
				sectionsList = append(sectionsList, currentSectionName)
			} else {
				currentSectionName = foundSectionName
			}
			continue
		}

		finalLine := line
		var ownersList []string
		terms := strings.Fields(line)
		for _, term := range terms {
			if len(term) == 0 {
				continue
			}

			// Identify owners by their prefixes (either @ for usernames or containing @ for emails)
			if term[0] == '@' || strings.Contains(term, "@") {
				ownersList = append(ownersList, term)
				pos := strings.Index(finalLine, term)
				if pos > 0 {
					finalLine = finalLine[:pos] + finalLine[pos+len(term):]
				}
			}
		}

		finalLine = strings.TrimSpace(finalLine)
		if len(finalLine) == 0 {
			continue
		}

		entriesList = append(entriesList, Entry{Pattern: finalLine, Owners: ownersList, Section: currentSectionName})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Reverse the entries list to maintain the order of appearance in the file
	for i, j := 0, len(entriesList)-1; i < j; i, j = i+1, j-1 {
		entriesList[i], entriesList[j] = entriesList[j], entriesList[i]
	}

	// Group entries by section
	sections := make(map[string][]Entry)
	for _, entry := range entriesList {
		sections[entry.Section] = append(sections[entry.Section], entry)
	}

	return &CodeOwners{Sections: sections}, nil
}

// findSectionIgnoreCase searches for a section name in a case-insensitive manner.
// It returns the section name if found, otherwise returns an empty string.
func findSectionIgnoreCase(sections []string, section string) string {
	sectionLower := strings.ToLower(section)
	for _, s := range sections {
		if strings.ToLower(s) == sectionLower {
			return s
		}
	}
	return ""
}

// Match finds the first entry in the CodeOwners that matches the given value.
// It returns a pointer to the matched entry, or nil if no match is found.
func (co *CodeOwners) Match(value string) *Entry {
	var matchedEntries []Entry

	for _, section := range co.Sections {
		for _, entry := range section {
			pattern := entry.Pattern
			finalPattern := pattern

			var includeAnythingBefore, includeAnythingAfter bool

			if strings.HasPrefix(pattern, "/") {
				includeAnythingBefore = false
			} else {
				if strings.HasPrefix(finalPattern, "*") {
					finalPattern = finalPattern[1:]
				}
				includeAnythingBefore = true
			}

			if strings.HasSuffix(pattern, "/") {
				includeAnythingAfter = true
			} else if strings.HasSuffix(pattern, "/*") {
				includeAnythingAfter = true
				finalPattern = finalPattern[:len(finalPattern)-1]
			} else {
				includeAnythingAfter = false
			}

			if includeAnythingAfter {
				found := includeAnythingBefore && strings.Contains(value, finalPattern) || strings.HasPrefix(value, finalPattern)
				if !found {
					continue
				}

				if !strings.HasSuffix(pattern, "/*") {
					matchedEntries = append(matchedEntries, entry)
					break
				}

				patternEnd := strings.Index(value, finalPattern)
				if patternEnd != -1 {
					patternEnd += len(finalPattern)
					remainingString := value[patternEnd:]
					if strings.Index(remainingString, "/") == -1 {
						matchedEntries = append(matchedEntries, entry)
						break
					}
				}
			} else {
				if includeAnythingBefore {
					if strings.HasSuffix(value, finalPattern) {
						matchedEntries = append(matchedEntries, entry)
						break
					}
				} else if value == finalPattern {
					matchedEntries = append(matchedEntries, entry)
					break
				}
			}
		}
	}

	switch len(matchedEntries) {
	case 0:
		return nil
	case 1:
		return &matchedEntries[0]
	default:
		patterns := make([]string, 0)
		owners := make([]string, 0)
		sections := make([]string, 0)
		for _, entry := range matchedEntries {
			patterns = append(patterns, entry.Pattern)
			owners = append(owners, entry.Owners...)
			sections = append(sections, entry.Section)
		}
		return &Entry{
			Pattern: strings.Join(patterns, " | "),
			Owners:  owners,
			Section: strings.Join(sections, " | "),
		}
	}
}

// GetOwnersString returns a formatted string of the owners list in an Entry.
// It returns an empty string if there are no owners.
func (e Entry) GetOwnersString() string {
	if e.Owners == nil || len(e.Owners) == 0 {
		return ""
	}

	return "[\"" + strings.Join(e.Owners, "\",\"") + "\"]"
}
