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
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type filesWithPatternStr struct {
	files   []checker.File
	pattern string
}
type languageFuzzConfig struct {
	fuzzFileMatchPattern, fuzzFuncRegexPattern, langFuzzDocumentURL, langFuzzDesc string
	//TODO: more language fuzzing-related fields
}

// Contains fuzzing speficications for programming languages.
// Use lowercases as the key, such as go, python, javascript, c++, etc.
var languageFuzzSpecsMap = map[string]languageFuzzConfig{
	// Default fuzz patterns for Go.
	"go": {
		fuzzFileMatchPattern: "*_test.go",
		fuzzFuncRegexPattern: `func\s+Fuzz\w+\s*\(\w+\s+\*testing.F\)`,
		langFuzzDocumentURL:  "https://go.dev/doc/fuzz/",
		langFuzzDesc:         "Go fuzzing intelligently walks through the code to find and report failures",
	},
	// TODO: add more language-speficic fuzz patterns & configs.
}

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

	usingFuzzFunc, files, e := checkFuzzFunc(c)
	if e != nil {
		return checker.FuzzingData{}, fmt.Errorf("%w", e)
	}
	if usingFuzzFunc {
		fuzzers = append(fuzzers,
			checker.Tool{
				Name: "user-defined fuzz functions",
				URL:  asPointer(languageFuzzSpecsMap["go"].langFuzzDocumentURL),
				Desc: asPointer(languageFuzzSpecsMap["go"].langFuzzDesc),
				File: &files[0],
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

func checkFuzzFunc(c *checker.CheckRequest) (bool, []checker.File, error) {
	if c.RepoClient == nil {
		return false, nil, nil
	}
	// Use GitHub API to decide the primary programming language to be checked for the repo.
	// Currently, we only perform the fuzzing check for a single language per repo.
	// TODO: maybe add multi-language fuzzing check for each repo.
	// languageToBeChecked := c.Repo.Languages()
	languageToBeChecked := "go"
	languagePat, found := languageFuzzSpecsMap[languageToBeChecked]
	if !found {
		return false, nil, fmt.Errorf("current repo language %s not supported", languageToBeChecked)
	}
	filePattern, funcPattern := languagePat.fuzzFileMatchPattern, languagePat.fuzzFuncRegexPattern
	if filePattern == "" || funcPattern == "" {
		return false, nil, fmt.Errorf("file/func fuzz patterns not found for %s", languageToBeChecked)
	}

	matcher := fileparser.PathMatcher{
		Pattern:       filePattern,
		CaseSensitive: false,
	}

	data := filesWithPatternStr{
		files:   make([]checker.File, 0),
		pattern: funcPattern,
	}
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, matcher, getFuzzFunc, &data)
	if err != nil {
		return false, nil, fmt.Errorf("error when OnMatchingFileContentDo: %w", err)
	}
	return true, data.files, nil
}

// This is the callback func for interface OnMatchingFileContentDo,
// used for matching fuzz functions in the file content
// and return a list of files (or nil for not found).
var getFuzzFunc fileparser.DoWhileTrueOnFileContent = func(path string, content []byte, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("getFuzzFunc requires exactly one argument: %w", errInvalidArgLength)
	}
	pdata, ok := args[0].(*filesWithPatternStr)
	if !ok {
		return false, fmt.Errorf("invalid arg type: %w", errInvalidArgType)
	}
	r, _ := regexp.Compile(pdata.pattern)
	lines := strings.Split(string(content), `\n`)

	for i, line := range lines {
		found := r.FindString(line)
		if found != "" {
			pdata.files = append(pdata.files, checker.File{
				Path:    path,
				Type:    checker.FileTypeSource,
				Snippet: found,
				Offset:  uint(i + 1),
			})
		}
	}
	return true, nil
}
