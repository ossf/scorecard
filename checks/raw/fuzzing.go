// Copyright 2021 OpenSSF Scorecard Authors
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
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

const (
	fuzzerOSSFuzz                 = "OSSFuzz"
	fuzzerClusterFuzzLite         = "ClusterFuzzLite"
	fuzzerBuiltInGo               = "GoBuiltInFuzzer"
	fuzzerPropertyBasedHaskell    = "HaskellPropertyBasedTesting"
	fuzzerPropertyBasedJavaScript = "JavaScriptPropertyBasedTesting"
	fuzzerPropertyBasedTypeScript = "TypeScriptPropertyBasedTesting"
	fuzzerPythonAtheris           = "PythonAtherisFuzzer"
	fuzzerCLibFuzzer              = "CLibFuzzer"
	fuzzerCppLibFuzzer            = "CppLibFuzzer"
	fuzzerSwiftLibFuzzer          = "SwiftLibFuzzer"
	fuzzerRustCargoFuzz           = "RustCargoFuzzer"
	fuzzerJavaJazzerFuzzer        = "JavaJazzerFuzzer"
	// TODO: add more fuzzing check supports.
)

type filesWithPatternStr struct {
	pattern string
	files   []checker.File
}

// Configurations for language-specified fuzzers.
type languageFuzzConfig struct {
	URL, Desc *string

	funcPattern, Name string
	// TODO: add more language fuzzing-related fields.

	// Patterns are according to path.Match.
	filePatterns []string
}

// Contains fuzzing speficications for programming languages.
// Please use the type Language defined in clients/languages.go rather than a raw string.
var languageFuzzSpecs = map[clients.LanguageName]languageFuzzConfig{
	// Default fuzz patterns for Go.
	clients.Go: {
		filePatterns: []string{"*_test.go"},
		funcPattern:  `func\s+Fuzz\w+\s*\(\w+\s+\*testing.F\)`,
		Name:         fuzzerBuiltInGo,
		URL:          asPointer("https://go.dev/doc/fuzz/"),
		Desc: asPointer(
			"Go fuzzing intelligently walks through the source code to report failures and find vulnerabilities."),
	},
	// Fuzz patterns for Haskell based on property-based testing.
	//
	// Based on the import of one of these packages:
	// * https://hackage.haskell.org/package/QuickCheck
	// * https://hedgehog.qa/
	// * https://github.com/NorfairKing/validity
	// * https://hackage.haskell.org/package/smallcheck
	//
	// They can also be imported indirectly through these test frameworks:
	// * https://hspec.github.io/
	// * https://hackage.haskell.org/package/tasty
	//
	// This is not an exhaustive list.
	clients.Haskell: {
		filePatterns: []string{"*.hs", "*.lhs"},
		// Look for direct imports of QuickCheck, Hedgehog, validity, or SmallCheck,
		// or their indirect imports through the higher-level Hspec or Tasty testing frameworks.
		funcPattern: `import\s+(qualified\s+)?Test\.((Hspec|Tasty)\.)?(QuickCheck|Hedgehog|Validity|SmallCheck)`,
		Name:        fuzzerPropertyBasedHaskell,
		Desc:        propertyBasedDescription("Haskell"),
	},
	// Fuzz patterns for JavaScript and TypeScript based on property-based testing.
	//
	// Based on the import of one of these packages:
	// * https://github.com/dubzzz/fast-check/tree/main/packages/fast-check#readme
	// * https://github.com/dubzzz/fast-check/tree/main/packages/ava#readme
	// * https://github.com/dubzzz/fast-check/tree/main/packages/jest#readme
	// * https://github.com/dubzzz/fast-check/tree/main/packages/vitest#readme
	//
	// This is not an exhaustive list.
	clients.JavaScript: {
		filePatterns: []string{"*.js"},
		// Look for direct imports of fast-check and its test runners integrations.
		funcPattern: `(from\s+['"](fast-check|@fast-check/(ava|jest|vitest))['"]|` +
			`require\(\s*['"](fast-check|@fast-check/(ava|jest|vitest))['"]\s*\))`,
		Name: fuzzerPropertyBasedJavaScript,
		Desc: propertyBasedDescription("JavaScript"),
	},
	clients.TypeScript: {
		filePatterns: []string{"*.ts"},
		// Look for direct imports of fast-check and its test runners integrations.
		funcPattern: `(from\s+['"](fast-check|@fast-check/(ava|jest|vitest))['"]|` +
			`require\(\s*['"](fast-check|@fast-check/(ava|jest|vitest))['"]\s*\))`,
		Name: fuzzerPropertyBasedTypeScript,
		Desc: propertyBasedDescription("TypeScript"),
	},
	clients.Python: {
		filePatterns: []string{"*.py"},
		funcPattern:  `import atheris`,
		Name:         fuzzerPythonAtheris,
		Desc: asPointer(
			"Python fuzzing by way of Atheris"),
	},
	clients.C: {
		filePatterns: []string{"*.c"},
		funcPattern:  `LLVMFuzzerTestOneInput`,
		Name:         fuzzerCLibFuzzer,
		Desc: asPointer(
			"Fuzzed with C LibFuzzer"),
	},
	clients.Cpp: {
		filePatterns: []string{"*.cc", "*.cpp"},
		funcPattern:  `LLVMFuzzerTestOneInput`,
		Name:         fuzzerCppLibFuzzer,
		Desc: asPointer(
			"Fuzzed with cpp LibFuzzer"),
	},
	clients.Rust: {
		filePatterns: []string{"*.rs"},
		funcPattern:  `libfuzzer_sys`,
		Name:         fuzzerRustCargoFuzz,
		Desc: asPointer(
			"Fuzzed with Cargo-fuzz"),
	},
	clients.Java: {
		filePatterns: []string{"*.java"},
		funcPattern:  `com.code_intelligence.jazzer.api.FuzzedDataProvider;`,
		Name:         fuzzerJavaJazzerFuzzer,
		Desc: asPointer(
			"Fuzzed with Jazzer fuzzer"),
	},
	clients.Swift: {
		filePatterns: []string{"*.swift"},
		funcPattern:  `LLVMFuzzerTestOneInput`,
		Name:         fuzzerSwiftLibFuzzer,
		Desc: asPointer(
			"Fuzzed with Swift LibFuzzer"),
	},
	// TODO: add more language-specific fuzz patterns & configs.
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
				Name: fuzzerClusterFuzzLite,
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
				Name: fuzzerOSSFuzz,
				URL:  asPointer("https://github.com/google/oss-fuzz"),
				Desc: asPointer("Continuous Fuzzing for Open Source Software"),
				// TODO: File.
			},
		)
	}

	langs, err := c.RepoClient.ListProgrammingLanguages()
	if err != nil {
		return checker.FuzzingData{}, fmt.Errorf("cannot get langs of repo: %w", err)
	}
	prominentLangs := getProminentLanguages(langs)
	for _, lang := range prominentLangs {
		usingFuzzFunc, files, e := checkFuzzFunc(c, lang)
		if e != nil {
			return checker.FuzzingData{}, fmt.Errorf("%w", e)
		}
		if usingFuzzFunc {
			fuzzers = append(fuzzers,
				checker.Tool{
					Name:  languageFuzzSpecs[lang].Name,
					URL:   languageFuzzSpecs[lang].URL,
					Desc:  languageFuzzSpecs[lang].Desc,
					Files: files,
				},
			)
		}
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

func checkFuzzFunc(c *checker.CheckRequest, lang clients.LanguageName) (bool, []checker.File, error) {
	if c.RepoClient == nil {
		return false, nil, nil
	}
	data := filesWithPatternStr{
		files: make([]checker.File, 0),
	}
	// Search language-specified fuzz func patterns in the hashmap.
	pattern, found := languageFuzzSpecs[lang]
	if !found {
		// If the fuzz patterns for the current language not supported yet,
		// we return it as false (not found), nil (no files), and nil (no errors).
		return false, nil, nil
	}
	// Get patterns for file and func.
	// We use the file pattern in the matcher to match the test files,
	// and put the func pattern in var data to match file contents (func names).
	filePatterns, funcPattern := pattern.filePatterns, pattern.funcPattern
	var dataFiles []checker.File
	for _, filePattern := range filePatterns {
		matcher := fileparser.PathMatcher{
			Pattern:       filePattern,
			CaseSensitive: false,
		}
		data.pattern = funcPattern
		err := fileparser.OnMatchingFileContentDo(c.RepoClient, matcher, getFuzzFunc, &data)
		if err != nil {
			return false, nil, fmt.Errorf("error when OnMatchingFileContentDo: %w", err)
		}
		dataFiles = append(dataFiles, data.files...)
	}

	if len(dataFiles) == 0 {
		// This means no fuzz funcs matched for this language.
		return false, nil, nil
	}
	return true, dataFiles, nil
}

// This is the callback func for interface OnMatchingFileContentDo
// used for matching fuzz functions in the file content,
// and return a list of files (or nil for not found).
var getFuzzFunc fileparser.DoWhileTrueOnFileContent = func(
	path string, content []byte, args ...interface{},
) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("getFuzzFunc requires exactly one argument: %w", errInvalidArgLength)
	}
	pdata, ok := args[0].(*filesWithPatternStr)
	if !ok {
		return false, errInvalidArgType
	}
	r := regexp.MustCompile(pdata.pattern)
	lines := bytes.Split(content, []byte("\n"))
	for i, line := range lines {
		found := r.FindString(string(line))
		if found != "" {
			// If fuzz func is found in the file, add it to the file array,
			// with its file path as Path, func name as Snippet,
			// FileTypeFuzz as Type, and # of lines as Offset.
			pdata.files = append(pdata.files, checker.File{
				Path:    path,
				Type:    finding.FileTypeSource,
				Snippet: found,
				Offset:  uint(i + 1), // Since the # of lines starts from zero.
			})
		}
	}
	return true, nil
}

func getProminentLanguages(langs []clients.Language) []clients.LanguageName {
	numLangs := len(langs)
	if numLangs == 0 {
		return nil
	}
	totalLoC := 0
	// Use a map to record languages and their lines of code to drop potential duplicates.
	langMap := map[clients.LanguageName]int{}
	for _, l := range langs {
		totalLoC += l.NumLines
		langMap[l.Name] += l.NumLines
	}
	// Calculate the average lines of code in the current repo.
	// This var can stay as an int, no need for a precise float value.
	avgLoC := totalLoC / numLangs
	// Languages that have lines of code above average will be considered prominent.
	prominentThreshold := avgLoC / 4.0
	ret := []clients.LanguageName{}
	for lName, loC := range langMap {
		if loC >= prominentThreshold {
			lang := clients.LanguageName(strings.ToLower(string(lName)))
			ret = append(ret, lang)
		}
	}
	return ret
}

func propertyBasedDescription(language string) *string {
	s := fmt.Sprintf("Property-based testing in %s generates test instances randomly or exhaustively "+
		"and test that specific properties are satisfied.", language)
	return &s
}
