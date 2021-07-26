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
	"github.com/ossf/scorecard/checker"
)

// CheckIfFileExists downloads the tar of the repository and calls the onFile() to check
// for the occurrence.
func CheckIfFileExists(checkName string, c *checker.CheckRequest, onFile func(name string,
	dl checker.DetailLogger) (bool, error)) (bool, error) {
	matchedFiles, err := c.RepoClient.ListFiles(func(string) (bool, error) { return true, nil })
	if err != nil {
		// nolint: wrapcheck
		return false, err
	}
	for _, filename := range matchedFiles {
		rr, err := onFile(filename, c.Dlogger)
		if err != nil {
			return false, err
		}

		if rr {
			return true, nil
		}
	}

	return false, nil
}
