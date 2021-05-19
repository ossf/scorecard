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
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/ossf/scorecard/checker"
)

var ErrReadFile = errors.New("could not read entire file")

// IsMatchingPath uses 'pattern' to shell-match the 'path' and its filename
// 'caseSensitive' indicates the match should be case-sensitive. Default: no.
func IsMatchingPath(pattern, fullpath string, caseSensitive bool) (bool, error) {
	if !caseSensitive {
		pattern = strings.ToLower(pattern)
		fullpath = strings.ToLower(fullpath)
	}

	filename := path.Base(fullpath)
	match, err := path.Match(pattern, fullpath)
	switch {
	case err != nil:
		return false, fmt.Errorf("match error: %w", err)
	case !match:
		if match, err = path.Match(pattern, filename); err != nil || !match {
			return false, fmt.Errorf("match error: %w", err)
		}
	}

	return true, nil
}

func HeaderSizeMatchesFileSize(hdr *tar.Header, size int) bool {
	if hdr.Format != tar.FormatUSTAR &&
		hdr.Format != tar.FormatUnknown &&
		int64(size) != hdr.Size {
		return false
	}
	return true
}

func NonEmptyRegularFile(hdr *tar.Header) bool {
	return hdr.Typeflag == tar.TypeReg && hdr.Size > 0
}

// CheckFilesContent downloads the tar of the repository and calls the onFileContent() function
// shellPathFnPattern is used for https://golang.org/pkg/path/#Match
// Warning: the pattern is used to match (1) the entire path AND (2) the filename alone. This means:
// 	- To scope the search to a directory, use "./dirname/*". Example, for the root directory,
// 		use "./*".
//	- A pattern such as "*mypatern*" will match files containing mypattern in *any* directory.
func CheckFilesContent(checkName, shellPathFnPattern string,
	caseSensitive bool,
	c *checker.CheckRequest,
	onFileContent func(path string, content []byte,
		Logf func(s string, f ...interface{})) (bool, error),
) checker.CheckResult {
	r, _, err := c.Client.Repositories.Get(c.Ctx, c.Owner, c.Repo)
	if err != nil {
		return checker.MakeRetryResult(checkName, err)
	}
	url := r.GetArchiveURL()
	url = strings.Replace(url, "{archive_format}", "tarball/", 1)
	url = strings.Replace(url, "{/ref}", r.GetDefaultBranch(), 1)

	// Using the http.get instead of the lib httpClient because
	// the default checker.HTTPClient caches everything in the memory and it causes oom.

	//https://securego.io/docs/rules/g107.html
	//nolint
	resp, err := http.Get(url)
	if err != nil {
		return checker.MakeRetryResult(checkName, err)
	}
	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return checker.MakeRetryResult(checkName, err)
	}
	tr := tar.NewReader(gz)
	res := true

	for {
		hdr, err := tr.Next()
		if err != nil && err != io.EOF {
			return checker.MakeRetryResult(checkName, err)
		}

		if err == io.EOF {
			break
		}

		// Only consider regular files.
		if !NonEmptyRegularFile(hdr) {
			continue
		}

		// Strip the repo name
		const splitLength = 2
		names := strings.SplitN(hdr.Name, "/", splitLength)
		if len(names) < splitLength {
			continue
		}

		fullpath := names[1]

		// Filter out files based on path/names.
		b, err := IsMatchingPath(shellPathFnPattern, fullpath, caseSensitive)
		switch {
		case err != nil:
			return checker.MakeFailResult(checkName, err)
		case !b:
			continue
		}

		content := make([]byte, hdr.Size)
		n, err := tr.Read(content)
		if err != nil && err != io.EOF {
			return checker.MakeRetryResult(checkName, err)
		}
		// We should have reached the end of files AND
		// the number of bytes should be the same as number
		// indicated in header, unless the file format supports
		// sparse regions. Only USTAR format does not support
		// spare regions -- see https://golang.org/pkg/archive/tar/
		if b := HeaderSizeMatchesFileSize(hdr, n); !b {
			return checker.MakeRetryResult(checkName, ErrReadFile)
		}

		// We truncate the file to remove tailing 0 (sparse format).
		rr, err := onFileContent(fullpath, content[:n], c.Logf)
		if err != nil {
			return checker.MakeFailResult(checkName, err)
		}
		// We don't return rightway to give the onFileContent()
		// handler to log.
		if !rr {
			res = false
		}
	}

	if res {
		return checker.MakePassResult(checkName)
	}

	return checker.MakeFailResult(checkName, err)
}
