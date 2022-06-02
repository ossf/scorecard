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
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	FuzzNameOSSFuzz         = "OSS-Fuzz"
	FuzzNameClusterFuzzLite = "ClusterFuzzLite"
	FuzzNameUserDefinedFunc = "user-defined fuzz functions"
	// TODO: add more fuzz check support.
	NoFuzzCov  = 0.0 // No fuzz coverage, so the ratio is zero.
	AllFuzzCov = 1.0
)

type filesWithPatternStr struct {
	files   []checker.File
	pattern string
}
type languageFuzzConfig struct {
	fuzzFileMatchPattern, fuzzFuncRegexPattern, langFuzzDocumentURL, langFuzzDesc string
	//TODO: add more language fuzzing-related fields.
}

// Contains fuzzing speficications for programming languages.
// Use lowercases as the key, such as go, python, javascript, c++, etc.
var languageFuzzSpecsMap = map[string]languageFuzzConfig{
	// Default fuzz patterns for Go.
	"go": {
		fuzzFileMatchPattern: "*_test.go",
		fuzzFuncRegexPattern: `func\s+Fuzz\w+\s*\(\w+\s+\*testing.F\)`,
		langFuzzDocumentURL:  *asPointer("https://go.dev/doc/fuzz/"),
		langFuzzDesc:         *asPointer("Go fuzzing intelligently walks through the source code to report failures and find vulnerabilities."),
	},
	"python": {
		fuzzFileMatchPattern: "*_test.py",
		fuzzFuncRegexPattern: `func\s+Fuzz\w+\s*\(\w+\s+\*testing.F\)`,
		langFuzzDocumentURL:  *asPointer("py"),
		langFuzzDesc:         *asPointer("pypy"),
	},
	"javascript": {
		fuzzFileMatchPattern: "*_test.js",
		fuzzFuncRegexPattern: `func\s+Fuzz\w+\s*\(\w+\s+\*testing.F\)`,
		langFuzzDocumentURL:  *asPointer("js"),
		langFuzzDesc:         *asPointer("jsjs"),
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
				Name: FuzzNameClusterFuzzLite,
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
				Name: FuzzNameOSSFuzz,
				URL:  asPointer("https://github.com/google/oss-fuzz"),
				Desc: asPointer("Continuous Fuzzing for Open Source Software"),
				// TODO: File.
			},
		)
	}

	usingFuzzFunc, langCov, files, e := checkFuzzFunc(c)
	if e != nil {
		return checker.FuzzingData{}, fmt.Errorf("%w", e)
	}
	if usingFuzzFunc {
		fuzzers = append(fuzzers,
			checker.Tool{
				Name:             FuzzNameUserDefinedFunc,
				URL:              asPointer(languageFuzzSpecsMap["go"].langFuzzDocumentURL),
				Desc:             asPointer(languageFuzzSpecsMap["go"].langFuzzDesc),
				File:             files,
				LanguageCoverage: langCov,
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

func checkFuzzFunc(c *checker.CheckRequest) (bool, float32, []checker.File, error) {
	if c.RepoClient == nil {
		return false, NoFuzzCov, nil, fmt.Errorf("empty RepoClient")
	}
	// To get the prominent programming language(s) to be checked.
	langMap, err := c.RepoClient.ListProgrammingLanguages()
	if err != nil {
		return false, NoFuzzCov, nil, fmt.Errorf("get programming languages of repo failed %w", err)
	}
	langsProminent, err := getProminentLanguages(langMap)
	if err != nil {
		return false, NoFuzzCov, nil, fmt.Errorf("error when getting promiment languages: %w", err)
	}
	fmt.Println(*langsProminent) // For debug.

	data := filesWithPatternStr{
		files: make([]checker.File, 0),
	}
	fuzzed, notFuzzed := &[]string{}, &[]string{}

	// Iterate the prominant language list and check for fuzz funcs per language.
	for _, lang := range *langsProminent {
		// Search language fuzz patterns in the hashmap.
		pattern, found := languageFuzzSpecsMap[lang]
		if !found {
			log.Printf("fuzz patterns for the current language \"%s\" not supported", lang)
			continue
		}
		// Get patterns for file and func.
		filePattern, funcPattern := pattern.fuzzFileMatchPattern, pattern.fuzzFuncRegexPattern
		matcher := fileparser.PathMatcher{
			Pattern:       filePattern,
			CaseSensitive: false,
		}
		data.pattern = funcPattern
		oldFilesLen := len(data.files) // Files length before checking.
		err = fileparser.OnMatchingFileContentDo(c.RepoClient, matcher, getFuzzFunc, &data)
		if err != nil {
			return false, NoFuzzCov, nil, fmt.Errorf("error when OnMatchingFileContentDo: %w", err)
		}
		if len(data.files) == oldFilesLen {
			// If the files length after checking doesn't increase after checking,
			// it indicates no fuzz funcs found for the current language so we give it a false.
			*notFuzzed = append(*notFuzzed, lang)
		} else {
			// Meaning the current lang is fuzzed.
			*fuzzed = append(*fuzzed, lang)
		}
	}
	// Calculate the fuzz coverage ratio for prominent languages.
	l1, l2 := len(*fuzzed), len(*notFuzzed)
	langFuzzCov := float32(l1) / (float32(l1) + float32(l2))
	if langFuzzCov != AllFuzzCov {
		log.Printf("not all prominent languages are fuzzed")
		log.Printf("fuzzed lang: %s, not fuzzed lang: %s, language fuzz coverage: %.2f",
			*fuzzed, *notFuzzed, langFuzzCov)
	}
	if langFuzzCov == NoFuzzCov {
		// Although not all prominent languages are fuzzed, we still return the files
		// that have been matched for fuzzed languages.
		return false, NoFuzzCov, data.files, nil
	} else {
		// All the prominent languages are fuzz-covered.
		return true, langFuzzCov, data.files, nil
	}
}

// This is the callback func for interface OnMatchingFileContentDo
// used for matching fuzz functions in the file content,
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
	lines := bytes.Split(content, []byte("\n"))
	for i, line := range lines {
		found := r.FindString(string(line))
		if found != "" {
			// If fuzz func is found in the file, add it to the file array,
			// with its file path as Path, func name as Snippet,
			// FileTypeFuzz as Type, and # of lines as Offset.
			pdata.files = append(pdata.files, checker.File{
				Path:    path,
				Type:    checker.FileTypeSource,
				Snippet: found,
				Offset:  uint(i + 1), // Since the # of lines starts from zero.
			})
		}
	}
	return true, nil
}

func getProminentLanguages(langs map[string]int) (*[]string, error) {
	if langs == nil {
		return nil, fmt.Errorf("no languages found in map")
	}
	numLangs := len(langs)
	totalLoC := 0
	for _, LoC := range langs {
		totalLoC += LoC
		numLangs++
	}
	// Var avgLoC calculates the average lines of code in the current repo,
	// and it can stay as an int, no need for a float value.
	avgLoC := totalLoC / numLangs

	// Languages that has lines of code above average will be considered prominent.
	ret := &[]string{}
	for lang, LoC := range langs {
		if LoC >= avgLoC {
			*ret = append(*ret, strings.ToLower(lang))
		}
	}
	return ret, nil
}
