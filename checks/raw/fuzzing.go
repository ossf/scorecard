// Copyright 2021 Security Scorecard Authors
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

package raw

import (
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

// Fuzzing runs Fuzzing check.
func Fuzzing(c *checker.CheckRequest) (checker.FuzzingData, error) {
	var fuzzers []checker.Tool
	usingCFLite, e := checkCFLite(c)
	if e != nil {
		return checker.FuzzingData{}, fmt.Errorf("%w", e)
	}
	if usingCFLite {
		fuzzers = append(fuzzers,
			checker.Tool{
				Name: "ClusterFuzzLite",
				URL:  asPointer("https://github.com/google/clusterfuzzlite"),
				Desc: asPointer("continuous fuzzing solution that runs as part of Continuous Integration (CI) workflows"),
				// TODO: File.
			},
		)
	}

	usingOSSFuzz, e := checkOSSFuzz(c)
	if e != nil {
		return checker.FuzzingData{}, fmt.Errorf("%w", e)
	}
	if usingOSSFuzz {
		fuzzers = append(fuzzers,
			checker.Tool{
				Name: "OSS-Fuzz",
				URL:  asPointer("https://github.com/google/oss-fuzz"),
				Desc: asPointer("Continuous Fuzzing for Open Source Software"),
				// TODO: File.
			},
		)
	}

	return checker.FuzzingData{Fuzzers: fuzzers}, nil
}

func checkCFLite(c *checker.CheckRequest) (bool, error) {
	result := false
	e := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".clusterfuzzlite/Dockerfile",
		CaseSensitive: true,
	}, func(path string, content []byte, args ...interface{}) (bool, error) {
		result = fileparser.CheckFileContainsCommands(content, "#")
		return false, nil
	}, nil)
	if e != nil {
		return result, fmt.Errorf("%w", e)
	}

	return result, nil
}

func checkOSSFuzz(c *checker.CheckRequest) (bool, error) {
	if c.OssFuzzRepo == nil {
		return false, nil
	}

	req := clients.SearchRequest{
		Query:    c.RepoClient.URI(),
		Filename: "project.yaml",
	}
	result, err := c.OssFuzzRepo.Search(req)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Search.Code: %v", err))
		return false, e
	}
	return result.Hits > 0, nil
}
