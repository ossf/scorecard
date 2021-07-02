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
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/ossf/scorecard/checker"
)

// ErrReadFile indicates the header size does not match the size of the file.
var ErrReadFile = errors.New("could not read entire file")

// IsMatchingPath uses 'pattern' to shell-match the 'path' and its filename
// 'caseSensitive' indicates the match should be case-sensitive. Default: no.
func isMatchingPath(pattern, fullpath string, caseSensitive bool) (bool, error) {
	if !caseSensitive {
		pattern = strings.ToLower(pattern)
		fullpath = strings.ToLower(fullpath)
	}

	filename := path.Base(fullpath)
	match, err := path.Match(pattern, fullpath)
	if err != nil {
		return false, fmt.Errorf("match error: %w", err)
	}

	// No match on the fullpath, let's try on the filename only.
	if !match {
		if match, err = path.Match(pattern, filename); err != nil {
			return false, fmt.Errorf("match error: %w", err)
		}
	}

	return match, nil
}

func isScorecardTestFile(owner, repo, fullpath string) bool {
	// testdata/ or /some/dir/testdata/some/other
	return owner == "ossf" && repo == "scorecard" && (strings.HasPrefix(fullpath, "testdata/") ||
		strings.Contains(fullpath, "/testdata/"))
}

// CheckFilesContent downloads the tar of the repository and calls the onFileContent() function
// shellPathFnPattern is used for https://golang.org/pkg/path/#Match
// Warning: the pattern is used to match (1) the entire path AND (2) the filename alone. This means:
// 	- To scope the search to a directory, use "./dirname/*". Example, for the root directory,
// 		use "./*".
//	- A pattern such as "*mypatern*" will match files containing mypattern in *any* directory.
//nolint
func CheckFilesContent(checkName, shellPathFnPattern string,
	caseSensitive bool,
	c *checker.CheckRequest,
	onFileContent func(path string, content []byte,
		Logf func(s string, f ...interface{})) (bool, error),
) checker.CheckResult {
	predicate := func(filepath string) bool {
		// Filter out Scorecard's own test files.
		if isScorecardTestFile(c.Owner, c.Repo, filepath) {
			return false
		}
		// Filter out files based on path/names using the pattern.
		b, err := isMatchingPath(shellPathFnPattern, filepath, caseSensitive)
		if err != nil {
			c.Logf("error during isMatchingPath: %v", err)
			return false
		}
		return b
	}
	res := true
	for _, file := range c.RepoClient.ListFiles(predicate) {
		content, err := c.RepoClient.GetFileContent(file)
		if err != nil {
			return checker.MakeRetryResult(checkName, err)
		}

		rr, err := onFileContent(file, content, c.Logf)
		if err != nil {
			return checker.MakeFailResult(checkName, err)
		}
		// We don't return rightway to let the onFileContent()
		// handler log.
		if !rr {
			res = false
		}
	}
	if res {
		return checker.MakePassResult(checkName)
	}

	return checker.MakeFailResult(checkName, nil)
}

func CheckFilesContent2(shellPathFnPattern string,
	caseSensitive bool,
	c *checker.CheckRequest,
	onFileContent func(path string, content []byte,
		l checker.Logf2) (bool, error),
) (bool, error) {
	predicate := func(filepath string) bool {
		// Filter out Scorecard's own test files.
		if isScorecardTestFile(c.Owner, c.Repo, filepath) {
			return false
		}
		// Filter out files based on path/names using the pattern.
		b, err := isMatchingPath(shellPathFnPattern, filepath, caseSensitive)
		if err != nil {
			c.Logf2(checker.DetailWarning, "4567", "internal error: %v", err)
			return false
		}
		return b
	}
	res := true
	for _, file := range c.RepoClient.ListFiles(predicate) {
		content, err := c.RepoClient.GetFileContent(file)
		if err != nil {
			return false, err
		}

		rr, err := onFileContent(file, content, c.Logf2)
		if err != nil {
			return false, err
		}
		// We don't return rightway to let the onFileContent()
		// handler log.
		if !rr {
			res = false
		}
	}

	return res, nil
}
