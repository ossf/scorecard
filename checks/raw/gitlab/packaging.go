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

package gitlab

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/finding"
)

// Packaging checks for packages.
func Packaging(c *checker.CheckRequest) (checker.PackagingData, error) {
	var data checker.PackagingData
	matchedFiles, err := c.RepoClient.ListFiles(fileparser.IsGitlabWorkflowFile)
	if err != nil {
		return data, fmt.Errorf("RepoClient.ListFiles: %w", err)
	}

	for _, fp := range matchedFiles {
		fc, err := c.RepoClient.GetFileContent(fp)
		if err != nil {
			return data, fmt.Errorf("RepoClient.GetFileContent: %w", err)
		}

		file, found := isGitlabPackagingWorkflow(fc, fp)

		if found {
			data.Packages = append(data.Packages, checker.Package{
				Name: new(string),
				Job:  &checker.WorkflowJob{},
				File: &file,
				Msg:  nil,
				Runs: []checker.Run{{URL: c.Repo.URI()}},
			})
			return data, nil
		}
	}

	return data, nil
}

func StringPointer(s string) *string {
	return &s
}

func isGitlabPackagingWorkflow(fc []byte, fp string) (checker.File, bool) {
	var lineNumber uint = checker.OffsetDefault

	packagingStrings := []string{
		"docker push",
		"nuget push",
		"poetry publish",
		"twine upload",
	}

ParseLines:
	for idx, val := range strings.Split(string(fc[:]), "\n") {
		for _, element := range packagingStrings {
			if strings.Contains(val, element) {
				lineNumber = uint(idx + 1)
				break ParseLines
			}
		}
	}

	return checker.File{
		Path:   fp,
		Offset: lineNumber,
		Type:   finding.FileTypeSource,
	}, lineNumber != checker.OffsetDefault
}
