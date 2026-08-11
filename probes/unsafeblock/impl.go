// Copyright 2025 OpenSSF Scorecard Authors
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

package unsafeblock

import (
	"bytes"
	"embed"
	"fmt"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/dotnet/csproj"
	"github.com/ossf/scorecard/v5/internal/probes"
)

//go:embed *.yml
var fs embed.FS

const (
	Probe = "unsafeblock"
)

type languageMemoryCheckConfig struct {
	funcPointer func(client *checker.CheckRequest) ([]finding.Finding, error)
	Desc        string
}

var languageMemorySafeSpecs = map[clients.LanguageName]languageMemoryCheckConfig{
	clients.Go: {
		funcPointer: checkGoUnsafePackage,
		Desc:        "Check if Go code uses the unsafe package",
	},

	clients.CSharp: {
		funcPointer: checkDotnetAllowUnsafeBlocks,
		Desc:        "Check if C# code uses unsafe blocks",
	},

	clients.Java: {
		funcPointer: checkJavaUnsafeClass,
		Desc:        "Check if Java code uses the Unsafe class",
	},
}

func init() {
	probes.MustRegisterIndependent(Probe, Run)
}

func Run(raw *checker.CheckRequest) (found []finding.Finding, probeName string, err error) {
	repoLanguageChecks, err := getLanguageChecks(raw)
	if err != nil {
		return nil, Probe, err
	}
	findings := []finding.Finding{}
	for _, lang := range repoLanguageChecks {
		langFindings, err := lang.funcPointer(raw)
		if err != nil {
			return nil, Probe, fmt.Errorf("error while running function for language %s: %w", lang.Desc, err)
		}
		findings = append(findings, langFindings...)
	}
	var nonErrorFindings bool
	for _, f := range findings {
		if f.Outcome != finding.OutcomeError {
			nonErrorFindings = true
		}
	}
	// if we don't have any findings (ignoring OutcomeError), we think it's safe
	if !nonErrorFindings {
		found, err := finding.NewWith(fs, Probe,
			"All supported ecosystems do not declare or use unsafe code blocks", nil, finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *found)
	}
	return findings, Probe, nil
}

func getLanguageChecks(raw *checker.CheckRequest) ([]languageMemoryCheckConfig, error) {
	langs, err := raw.RepoClient.ListProgrammingLanguages()
	if err != nil {
		return nil, fmt.Errorf("cannot get langs of repo: %w", err)
	}
	if len(langs) == 1 && langs[0].Name == clients.All {
		return getAllLanguages(), nil
	}
	ret := []languageMemoryCheckConfig{}
	for _, language := range langs {
		if lang, ok := languageMemorySafeSpecs[clients.LanguageName(strings.ToLower(string(language.Name)))]; ok {
			ret = append(ret, lang)
		}
	}
	return ret, nil
}

func getAllLanguages() []languageMemoryCheckConfig {
	allLanguages := make([]languageMemoryCheckConfig, 0, len(languageMemorySafeSpecs))
	for l := range languageMemorySafeSpecs {
		allLanguages = append(allLanguages, languageMemorySafeSpecs[l])
	}
	return allLanguages
}

// Golang

func checkGoUnsafePackage(client *checker.CheckRequest) ([]finding.Finding, error) {
	findings := []finding.Finding{}
	if err := fileparser.OnMatchingFileContentDo(client.RepoClient, fileparser.PathMatcher{
		Pattern:       "*.go",
		CaseSensitive: false,
	}, goCodeUsesUnsafePackage, &findings); err != nil {
		return nil, err
	}

	return findings, nil
}

func goCodeUsesUnsafePackage(path string, content []byte, args ...interface{}) (bool, error) {
	findings, ok := args[0].(*[]finding.Finding)
	if !ok {
		// panic if it is not correct type
		panic(fmt.Sprintf("expected type findings, got %v", reflect.TypeOf(args[0])))
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ImportsOnly)
	if err != nil {
		found, err := finding.NewWith(fs, Probe, "malformed golang file", &finding.Location{
			Path: path,
		}, finding.OutcomeError)
		if err != nil {
			return false, fmt.Errorf("create finding: %w", err)
		}
		*findings = append(*findings, *found)
		return true, nil
	}
	for _, i := range f.Imports {
		if i.Path.Value == `"unsafe"` {
			lineStart := uint(fset.Position(i.Pos()).Line)
			found, err := finding.NewWith(fs, Probe,
				"Golang code uses the unsafe package", &finding.Location{
					Path: path, LineStart: &lineStart,
				}, finding.OutcomeTrue)
			if err != nil {
				return false, fmt.Errorf("create finding: %w", err)
			}
			*findings = append(*findings, *found)
		}
	}

	return true, nil
}

// CSharp

func checkDotnetAllowUnsafeBlocks(client *checker.CheckRequest) ([]finding.Finding, error) {
	findings := []finding.Finding{}

	if err := fileparser.OnMatchingFileContentDo(client.RepoClient, fileparser.PathMatcher{
		Pattern:       "*.csproj",
		CaseSensitive: false,
	}, csProjAllosUnsafeBlocks, &findings); err != nil {
		return nil, err
	}
	return findings, nil
}

func csProjAllosUnsafeBlocks(path string, content []byte, args ...interface{}) (bool, error) {
	findings, ok := args[0].(*[]finding.Finding)
	if !ok {
		// panic if it is not correct type
		panic(fmt.Sprintf("expected type findings, got %v", reflect.TypeOf(args[0])))
	}

	unsafe, err := csproj.IsAllowUnsafeBlocksEnabled(content)
	if err != nil {
		found, err := finding.NewWith(fs, Probe, "malformed csproj file", &finding.Location{
			Path: path,
		}, finding.OutcomeError)
		if err != nil {
			return false, fmt.Errorf("create finding: %w", err)
		}
		*findings = append(*findings, *found)
		return true, nil
	}
	if unsafe {
		found, err := finding.NewWith(fs, Probe,
			"C# project file allows the use of unsafe blocks", &finding.Location{
				Path: path,
			}, finding.OutcomeTrue)
		if err != nil {
			return false, fmt.Errorf("create finding: %w", err)
		}
		*findings = append(*findings, *found)
	}

	return true, nil
}

// Java

var (
	// javaMultiLineCommentRe matches /* ... */ comments (including across newlines).
	javaMultiLineCommentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)
	// javaSingleLineCommentRe matches // ... to end of line.
	javaSingleLineCommentRe = regexp.MustCompile(`//[^\n]*`)
	// javaMultiLineStringRe matches multi-line string literals.
	javaMultiLineStringRe = regexp.MustCompile(`(?s)""".*?"""`)
	// javaSingleLineStringRe matches single-line string literals.
	javaSingleLineStringRe = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)

	// javaUnsafeImportRe matches import statements for sun.misc.Unsafe or
	// jdk.internal.misc.Unsafe, allowing optional whitespace between tokens
	// (e.g. "import  sun . misc . Unsafe ;").
	javaUnsafeImportRe = regexp.MustCompile(
		`\bimport\s+(?:sun\s*\.\s*misc|jdk\s*\.\s*internal\s*\.\s*misc)\s*\.\s*Unsafe\s*;`)

	// javaUnsafeFQNRe matches fully-qualified references to sun.misc.Unsafe or
	// jdk.internal.misc.Unsafe in code (including inside import statements).
	javaUnsafeFQNRe = regexp.MustCompile(
		`\b(?:sun\s*\.\s*misc|jdk\s*\.\s*internal\s*\.\s*misc)\s*\.\s*Unsafe\b`)
)

// stripJavaComments removes single-line and multi-line comments from Java
// source, preserving newlines so that line numbers remain accurate.
func stripJavaComments(content []byte) []byte {
	// Replace multi-line comments with an equal number of newlines.
	src := javaMultiLineCommentRe.ReplaceAllFunc(content, func(match []byte) []byte {
		return bytes.Repeat([]byte("\n"), bytes.Count(match, []byte("\n")))
	})
	// Remove single-line comments (the newline itself is not part of the match).
	return javaSingleLineCommentRe.ReplaceAll(src, nil)
}

// stripJavaStringLiterals removes single-line and multi-line string literals from Java
// source, preserving newlines so that line numbers remain accurate.
func stripJavaStringLiterals(content []byte) []byte {
	// Replace multi-line string literals with an equal number of newlines.
	src := javaMultiLineStringRe.ReplaceAllFunc(content, func(match []byte) []byte {
		return bytes.Repeat([]byte("\n"), bytes.Count(match, []byte("\n")))
	})
	// Remove single-line string literals (the newline itself is not part of the match).
	return javaSingleLineStringRe.ReplaceAll(src, nil)
}

// javaLineNumber returns the 1-based line number of the byte at offset within src.
func javaLineNumber(src []byte, offset int) uint {
	return uint(bytes.Count(src[:offset], []byte("\n")) + 1)
}

func checkJavaUnsafeClass(client *checker.CheckRequest) ([]finding.Finding, error) {
	findings := []finding.Finding{}
	if err := fileparser.OnMatchingFileContentDo(client.RepoClient, fileparser.PathMatcher{
		Pattern:       "*.java",
		CaseSensitive: false,
	}, javaCodeUsesUnsafeClass, &findings); err != nil {
		return nil, err
	}

	return findings, nil
}

func javaCodeUsesUnsafeClass(path string, content []byte, args ...interface{}) (bool, error) {
	findings, ok := args[0].(*[]finding.Finding)
	if !ok {
		// panic if it is not correct type
		panic(fmt.Sprintf("expected type findings, got %v", reflect.TypeOf(args[0])))
	}

	src := stripJavaStringLiterals(stripJavaComments(content))

	// Report each import of an Unsafe class.
	importLocs := javaUnsafeImportRe.FindAllIndex(src, -1)
	for _, loc := range importLocs {
		line := javaLineNumber(src, loc[0])
		found, err := finding.NewWith(fs, Probe,
			"Java code imports the Unsafe class", &finding.Location{
				Path: path, LineStart: &line,
			}, finding.OutcomeTrue)
		if err != nil {
			return false, fmt.Errorf("create finding: %w", err)
		}
		*findings = append(*findings, *found)
	}

	// Report fully-qualified usages of an Unsafe class that are not part of
	// an import statement (i.e. direct usage without a prior import).
	for _, loc := range javaUnsafeFQNRe.FindAllIndex(src, -1) {
		if withinAnyRange(loc[0], importLocs) {
			continue
		}
		line := javaLineNumber(src, loc[0])
		found, err := finding.NewWith(fs, Probe,
			"Java code uses the Unsafe class", &finding.Location{
				Path: path, LineStart: &line,
			}, finding.OutcomeTrue)
		if err != nil {
			return false, fmt.Errorf("create finding: %w", err)
		}
		*findings = append(*findings, *found)
	}

	return true, nil
}

// withinAnyRange reports whether offset falls inside any of the given [start, end) ranges.
func withinAnyRange(offset int, ranges [][]int) bool {
	for _, r := range ranges {
		if offset >= r[0] && offset < r[1] {
			return true
		}
	}
	return false
}
