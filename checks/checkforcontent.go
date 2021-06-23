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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
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

func headerSizeMatchesFileSize(hdr *tar.Header, size int) bool {
	return hdr.Format == tar.FormatUSTAR ||
		hdr.Format == tar.FormatUnknown ||
		int64(size) == hdr.Size
}

func nonEmptyRegularFile(hdr *tar.Header) bool {
	return hdr.Typeflag == tar.TypeReg && hdr.Size > 0
}

func isScorecardTestFile(owner, repo, fullpath string) bool {
	// testdata/ or /some/dir/testdata/some/other
	return owner == "ossf" && repo == "scorecard" && (strings.HasPrefix(fullpath, "testdata/") ||
		strings.Contains(fullpath, "/testdata/"))
}

func extractFullpath(fn string) (string, bool) {
	const splitLength = 2
	names := strings.SplitN(fn, "/", splitLength)
	if len(names) < splitLength {
		return "", false
	}

	fullpath := names[1]
	return fullpath, true
}

func getTarReader(in io.Reader) (*tar.Reader, error) {
	gz, err := gzip.NewReader(in)
	if err != nil {
		return nil, fmt.Errorf("gzip reader failed: %w", err)
	}
	tr := tar.NewReader(gz)
	return tr, nil
}

func readEntireFile(tr *tar.Reader) (content []byte, err error) {
	var buf bytes.Buffer
	_, err = io.Copy(&buf, tr)
	if err != nil {
		return nil, fmt.Errorf("io.Copy: %w", err)
	}

	return buf.Bytes(), nil
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
	archiveReader, err := c.RepoClient.GetRepoArchiveReader()
	if err != nil {
		return checker.MakeRetryResult(checkName, err)
	}
	defer archiveReader.Close()
	tr, err := getTarReader(archiveReader)
	if err != nil {
		return checker.MakeRetryResult(checkName, err)
	}

	res := true

	var fullpath string
	var b bool
	for {
		hdr, err := tr.Next()
		if err != nil && err != io.EOF {
			return checker.MakeRetryResult(checkName, err)
		}

		if err == io.EOF {
			break
		}

		// Only consider regular files.
		if !nonEmptyRegularFile(hdr) {
			continue
		}

		// Extract the fullpath without the repo name.
		if fullpath, b = extractFullpath(hdr.Name); !b {
			continue
		}

		// Filter out Scorecard's own test files.
		if isScorecardTestFile(c.Owner, c.Repo, fullpath) {
			continue
		}

		// Filter out files based on path/names using the pattern.
		b, err := isMatchingPath(shellPathFnPattern, fullpath, caseSensitive)
		switch {
		case err != nil:
			return checker.MakeFailResult(checkName, err)
		case !b:
			continue
		}

		content, err := readEntireFile(tr)
		if err != nil {
			return checker.MakeRetryResult(checkName, err)
		}

		// We should have reached the end of files AND
		// the number of bytes should be the same as number
		// indicated in header, unless the file format supports
		// sparse regions. Only USTAR format does not support
		// spare regions -- see https://golang.org/pkg/archive/tar/
		if b := headerSizeMatchesFileSize(hdr, len(content)); !b {
			return checker.MakeRetryResult(checkName, ErrReadFile)
		}

		rr, err := onFileContent(fullpath, content, c.Logf)
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

	return checker.MakeFailResult(checkName, err)
}
