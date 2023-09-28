// Copyright 2020 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fileparser

import (
	"bufio"
	"fmt"
	"path"
	"strings"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

// isMatchingPath uses 'pattern' to shell-match the 'path' and its filename
// 'caseSensitive' indicates the match should be case-sensitive. Default: no.
func isMatchingPath(fullpath string, matchPathTo PathMatcher) (bool, error) {
	pattern := matchPathTo.Pattern
	if !matchPathTo.CaseSensitive {
		pattern = strings.ToLower(matchPathTo.Pattern)
		fullpath = strings.ToLower(fullpath)
	}

	filename := path.Base(fullpath)
	match, err := path.Match(pattern, fullpath)
	if err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalFilenameMatch, err))
	}

	// No match on the fullpath, let's try on the filename only.
	if !match {
		if match, err = path.Match(pattern, filename); err != nil {
			return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalFilenameMatch, err))
		}
	}

	return match, nil
}

func isTestdataFile(fullpath string) bool {
	// testdata/ or /some/dir/testdata/some/other
	return strings.HasPrefix(fullpath, "testdata/") ||
		strings.Contains(fullpath, "/testdata/") ||
		strings.HasPrefix(fullpath, "src/test/") ||
		strings.Contains(fullpath, "/src/test/")
}

// PathMatcher represents a query for a filepath.
type PathMatcher struct {
	Pattern       string
	CaseSensitive bool
}

// DoWhileTrueOnFileContent takes a filepath, its content and
// optional variadic args. It returns a boolean indicating whether
// iterating over next files should continue.
type DoWhileTrueOnFileContent func(path string, content []byte, args ...interface{}) (bool, error)

// OnMatchingFileContentDo matches all files listed by `repoClient` against `matchPathTo`
// and on every successful match, runs onFileContent fn on the file's contents.
// Continues iterating along the matched files until onFileContent returns
// either a false value or an error.
func OnMatchingFileContentDo(repoClient clients.RepoClient, matchPathTo PathMatcher,
	onFileContent DoWhileTrueOnFileContent, args ...interface{},
) error {
	predicate := func(filepath string) (bool, error) {
		// Filter out test files.
		if isTestdataFile(filepath) {
			return false, nil
		}
		// Filter out files based on path/names using the pattern.
		b, err := isMatchingPath(filepath, matchPathTo)
		if err != nil {
			return false, err
		}
		return b, nil
	}

	matchedFiles, err := repoClient.ListFiles(predicate)
	if err != nil {
		return fmt.Errorf("error during ListFiles: %w", err)
	}

	for _, file := range matchedFiles {
		content, err := repoClient.GetFileContent(file)
		if err != nil {
			return fmt.Errorf("error during GetFileContent: %w", err)
		}

		continueIter, err := onFileContent(file, content, args...)
		if err != nil {
			return err
		}

		if !continueIter {
			break
		}
	}

	return nil
}

// DoWhileTrueOnFilename takes a filename and optional variadic args and returns
// true if the next filename should continue to be processed.
type DoWhileTrueOnFilename func(path string, args ...interface{}) (bool, error)

// OnAllFilesDo iterates through all files returned by `repoClient` and
// calls `onFile` fn on them until `onFile` returns error or a false value.
func OnAllFilesDo(repoClient clients.RepoClient, onFile DoWhileTrueOnFilename, args ...interface{}) error {
	matchedFiles, err := repoClient.ListFiles(func(string) (bool, error) { return true, nil })
	if err != nil {
		return fmt.Errorf("error during ListFiles: %w", err)
	}
	for _, filename := range matchedFiles {
		continueIter, err := onFile(filename, args...)
		if err != nil {
			return err
		}

		if !continueIter {
			break
		}
	}
	return nil
}

// CheckFileContainsCommands checks if the file content contains commands or not.
// `comment` is the string or character that indicates a comment:
// for example for Dockerfiles, it would be `#`.
func CheckFileContainsCommands(content []byte, comment string) bool {
	if len(content) == 0 {
		return false
	}

	r := strings.NewReader(string(content))
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && !strings.HasPrefix(line, comment) {
			return true
		}
	}
	return false
}

// IsTemplateFile returns true if the file name contains a string commonly used in template files.
func IsTemplateFile(pathfn string) bool {
	parts := strings.FieldsFunc(path.Base(pathfn), func(r rune) bool {
		switch r {
		case '.', '-', '_':
			return true
		default:
			return false
		}
	})
	for _, part := range parts {
		switch strings.ToLower(part) {
		case "template", "tmpl", "tpl":
			return true
		}
	}
	return false
}
