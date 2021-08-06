// Copyright 2020 Security Scorecard Authors
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

package checks

import (
	"bufio"
	"fmt"
	"path"
	"strings"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// isMatchingPath uses 'pattern' to shell-match the 'path' and its filename
// 'caseSensitive' indicates the match should be case-sensitive. Default: no.
func isMatchingPath(pattern, fullpath string, caseSensitive bool) (bool, error) {
	if !caseSensitive {
		pattern = strings.ToLower(pattern)
		fullpath = strings.ToLower(fullpath)
	}

	filename := path.Base(fullpath)
	match, err := path.Match(pattern, fullpath)
	if err != nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalFilenameMatch, err))
	}

	// No match on the fullpath, let's try on the filename only.
	if !match {
		if match, err = path.Match(pattern, filename); err != nil {
			//nolint
			return false, sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalFilenameMatch, err))
		}
	}

	return match, nil
}

func isScorecardTestFile(owner, repo, fullpath string) bool {
	// testdata/ or /some/dir/testdata/some/other
	return owner == "ossf" && repo == "scorecard" && (strings.HasPrefix(fullpath, "testdata/") ||
		strings.Contains(fullpath, "/testdata/"))
}

// FileCbData is any data the caller can act upon
// to keep state.
type FileCbData interface{}

// FileContentCb is the callback.
// The bool returned indicates whether the CheckFilesContent2
// should continue iterating over files or not.
type FileContentCb func(path string, content []byte,
	dl checker.DetailLogger, data FileCbData) (bool, error)

// CheckFilesContent downloads the tar of the repository and calls the onFileContent() function
// shellPathFnPattern is used for https://golang.org/pkg/path/#Match
// Warning: the pattern is used to match (1) the entire path AND (2) the filename alone. This means:
// 	- To scope the search to a directory, use "./dirname/*". Example, for the root directory,
// 		use "./*".
//	- A pattern such as "*mypatern*" will match files containing mypattern in *any* directory.
func CheckFilesContent(shellPathFnPattern string,
	caseSensitive bool,
	c *checker.CheckRequest,
	onFileContent FileContentCb,
	data FileCbData,
) error {
	predicate := func(filepath string) (bool, error) {
		// Filter out Scorecard's own test files.
		if isScorecardTestFile(c.Owner, c.Repo, filepath) {
			return false, nil
		}
		// Filter out files based on path/names using the pattern.
		b, err := isMatchingPath(shellPathFnPattern, filepath, caseSensitive)
		if err != nil {
			return false, err
		}
		return b, nil
	}

	matchedFiles, err := c.RepoClient.ListFiles(predicate)
	if err != nil {
		// nolint: wrapcheck
		return err
	}

	for _, file := range matchedFiles {
		content, err := c.RepoClient.GetFileContent(file)
		if err != nil {
			//nolint
			return err
		}

		continueIter, err := onFileContent(file, content, c.Dlogger, data)
		if err != nil {
			return err
		}

		if !continueIter {
			break
		}
	}

	return nil
}

// FileCb represents a callback fn.
type FileCb func(path string,
	dl checker.DetailLogger, data FileCbData) (bool, error)

// CheckIfFileExists downloads the tar of the repository and calls the onFile() to check
// for the occurrence.
func CheckIfFileExists(checkName string, c *checker.CheckRequest, onFile FileCb, data FileCbData) error {
	matchedFiles, err := c.RepoClient.ListFiles(func(string) (bool, error) { return true, nil })
	if err != nil {
		// nolint: wrapcheck
		return err
	}
	for _, filename := range matchedFiles {
		continueIter, err := onFile(filename, c.Dlogger, data)
		if err != nil {
			return err
		}

		if !continueIter {
			break
		}
	}

	return nil
}

// FileGetCbDataAsBoolPointer returns callback data as bool.
func FileGetCbDataAsBoolPointer(data FileCbData) *bool {
	pdata, ok := data.(*bool)
	if !ok {
		// This never happens.
		panic("invalid type")
	}
	return pdata
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
