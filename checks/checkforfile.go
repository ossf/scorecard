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
	"io"
	"net/http"
	"strings"

	"github.com/ossf/scorecard/checker"
)

// CheckIfFileExists downloads the tar of the repository and calls the predicate to check
// for the occurrence.
func CheckIfFileExists(checkName string, c checker.CheckRequest, predicate func(name string,
	Logf func(s string, f ...interface{})) bool) checker.CheckResult {
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

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return checker.MakeRetryResult(checkName, err)
		}

		// Strip the repo name
		const splitLength = 2
		names := strings.SplitN(hdr.Name, "/", splitLength)
		if len(names) < splitLength {
			continue
		}

		name := names[1]
		if predicate(name, c.Logf) {
			return checker.MakePassResult(checkName)
		}
	}
	const confidence = 5
	return checker.CheckResult{
		Name:       checkName,
		Pass:       false,
		Confidence: confidence,
	}
}
