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
	"io"
	"strings"

	"github.com/ossf/scorecard/checker"
)

// CheckIfFileExists downloads the tar of the repository and calls the onFile() to check
// for the occurrence.
func CheckIfFileExists(checkName string, c *checker.CheckRequest, onFile func(name string,
	Logf func(s string, f ...interface{})) (bool, error)) checker.CheckResult {
	archiveReader, err := c.RepoClient.GetRepoArchiveReader()
	if err != nil {
		return checker.MakeRetryResult(checkName, err)
	}
	defer archiveReader.Close()
	gz, err := gzip.NewReader(archiveReader)
	if err != nil {
		return checker.MakeRetryResult(checkName, err)
	}
	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
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
		rr, err := onFile(name, c.Logf)
		if err != nil {
			return checker.CheckResult{
				Name:       checkName,
				Pass:       false,
				Confidence: checker.MaxResultConfidence,
				Error:      err,
			}
		}

		if rr {
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
