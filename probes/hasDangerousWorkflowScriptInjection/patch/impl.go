// Copyright 2024 OpenSSF Scorecard Authors
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

/*
TODO
  - Detects the end of the existing envvars at the first line that does not declare an
    envvar. This can lead to weird insertion positions if there is a comment in the
    middle of the `env:` block.
  - Tried performing a "dumber" implementation than the Python script, with less
    "parsing" of the workflow. However, the location given by f.Offset isn't precise
    enough. It only marks the start of the `run:` command, not the line where the
    variable is actually used. Will therefore need to, at least, parse the `run`
    command to replace all the instances of the unsafe variable. This means we can
    have multiple identical remediations if the same variable is used multiple times
    in the same step... that's just life.
*/
package patch

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/ossf/scorecard/v5/checker"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

const (
	assumedIndent = 2
)

func GeneratePatch(f checker.File, content string) string {
	unsafeVar := strings.Trim(f.Snippet, " ")
	runCmdIndex := f.Offset - 1

	lines := strings.Split(string(content), "\n")

	unsafePattern, envvar, ok := getReplacementRegexAndEnvvarName(unsafeVar)
	if !ok {
		return ""
	}

	replaceUnsafeVarWithEnvvar(lines, unsafePattern, envvar, runCmdIndex)

	lines, ok = addEnvvarsToGlobalEnv(lines, envvar, unsafeVar)
	if !ok {
		return ""
	}

	fixedWorkflow := strings.Join(lines, "\n")

	return getDiff(f.Path, content, fixedWorkflow)
}

func addEnvvarsToGlobalEnv(lines []string, envvar string, unsafeVar string) ([]string, bool) {
	globalIndentation, ok := findGlobalIndentation(lines)

	if !ok {
		// invalid workflow, could not determine global indentation
		return nil, false
	}

	envPos, envvarIndent, exists := findExistingEnv(lines, globalIndentation)

	if !exists {
		lines, envPos, ok = addNewGlobalEnv(lines, globalIndentation)
		if !ok {
			return nil, ok
		}

		// position now points to `env:`, insert variables below it
		envPos += 1
		envvarIndent = globalIndentation + assumedIndent
	}
	envvarDefinition := fmt.Sprintf("%s: ${{ %s }}", envvar, unsafeVar)
	lines = slices.Insert(lines, envPos,
		strings.Repeat(" ", envvarIndent)+envvarDefinition)
	return lines, ok
}

func addNewGlobalEnv(lines []string, globalIndentation int) ([]string, int, bool) {
	envPos, ok := findNewEnvPos(lines, globalIndentation)

	if !ok {
		// invalid workflow, could not determine location for new environment
		return nil, envPos, ok
	}

	label := strings.Repeat(" ", globalIndentation) + "env:"
	lines = slices.Insert(lines, envPos, []string{label, ""}...)
	return lines, envPos, ok
}

func findGlobalIndentation(lines []string) (int, bool) {
	r := regexp.MustCompile(`^\W*on:`)
	for _, line := range lines {
		if r.MatchString(line) {
			return getIndent(line), true
		}
	}

	return -1, false
}

func findExistingEnv(lines []string, globalIndent int) (int, int, bool) {
	num_lines := len(lines)
	indent := strings.Repeat(" ", globalIndent)

	// regex to detect the global `env:` block
	labelRegex := regexp.MustCompile(indent + "env:")
	i := 0
	for i = 0; i < num_lines; i++ {
		line := lines[i]
		if labelRegex.MatchString(line) {
			break
		}
	}

	if i >= num_lines-1 {
		// there must be at least one more line
		return -1, -1, false
	}

	i++ // move to line after `env:`
	envvarIndent := getIndent(lines[i])
	// regex to detect envvars belonging to the global `env:` block
	envvarRegex := regexp.MustCompile(indent + `\W+[^#]`)
	for ; i < num_lines; i++ {
		line := lines[i]
		if !envvarRegex.MatchString(line) {
			// no longer declaring envvars
			break
		}
	}

	return i, envvarIndent, true
}

func findNewEnvPos(lines []string, globalIndent int) (int, bool) {
	// the new env is added right before `jobs:`
	indent := strings.Repeat(" ", globalIndent)
	r := regexp.MustCompile(indent + "jobs:")
	for i, line := range lines {
		if r.MatchString(line) {
			return i, true
		}
	}

	return -1, false
}

type unsafePattern struct {
	envvarName   string
	idRegex      *regexp.Regexp
	replaceRegex *regexp.Regexp
}

func newUnsafePattern(e, p string) unsafePattern {
	return unsafePattern{
		envvarName:   e,
		idRegex:      regexp.MustCompile(p),
		replaceRegex: regexp.MustCompile(`{{\W*.*?` + p + `.*?\W*}}`),
	}
}

var unsafePatterns = []unsafePattern{
	newUnsafePattern("AUTHOR_EMAIL", `github\.event\.commits.*?\.author\.email`),
	newUnsafePattern("AUTHOR_EMAIL", `github\.event\.head_commit\.author\.email`),
	newUnsafePattern("AUTHOR_NAME", `github\.event\.commits.*?\.author\.name`),
	newUnsafePattern("AUTHOR_NAME", `github\.event\.head_commit\.author\.name`),
	newUnsafePattern("COMMENT_BODY", `github\.event\.comment\.body`),
	newUnsafePattern("COMMIT_MESSAGE", `github\.event\.commits.*?\.message`),
	newUnsafePattern("COMMIT_MESSAGE", `github\.event\.head_commit\.message`),
	newUnsafePattern("ISSUE_BODY", `github\.event\.issue\.body`),
	newUnsafePattern("ISSUE_TITLE", `github\.event\.issue\.title`),
	newUnsafePattern("PAGE_NAME", `github\.event\.pages.*?\.page_name`),
	newUnsafePattern("PR_BODY", `github\.event\.pull_request\.body`),
	newUnsafePattern("PR_DEFAULT_BRANCH", `github\.event\.pull_request\.head\.repo\.default_branch`),
	newUnsafePattern("PR_HEAD_LABEL", `github\.event\.pull_request\.head\.label`),
	newUnsafePattern("PR_HEAD_REF", `github\.event\.pull_request\.head\.ref`),
	newUnsafePattern("PR_TITLE", `github\.event\.pull_request\.title`),
	newUnsafePattern("REVIEW_BODY", `github\.event\.review\.body`),
	newUnsafePattern("REVIEW_COMMENT_BODY", `github\.event\.review_comment\.body`),

	newUnsafePattern("HEAD_REF", `github\.head_ref`),
}

func getReplacementRegexAndEnvvarName(unsafeVar string) (*regexp.Regexp, string, bool) {
	for _, p := range unsafePatterns {
		if p.idRegex.MatchString(unsafeVar) {
			return p.replaceRegex, p.envvarName, true
		}
	}
	return nil, "", false
}

func replaceUnsafeVarWithEnvvar(lines []string, replaceRegex *regexp.Regexp, envvar string, runIndex uint) {
	runIndent := getIndent(lines[runIndex])
	for i := int(runIndex); i < len(lines) && isParentLevelIndent(lines[i], runIndent); i++ {
		lines[i] = replaceRegex.ReplaceAllString(lines[i], envvar)
	}
}

func getIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " -"))
}

func isBlankOrComment(line string) bool {
	blank := regexp.MustCompile(`^\W*$`)
	comment := regexp.MustCompile(`^\W*#`)

	return blank.MatchString(line) || comment.MatchString(line)
}

func isParentLevelIndent(line string, parentIndent int) bool {
	if isBlankOrComment(line) {
		return false
	}
	return getIndent(line) >= parentIndent
}

// gets the changes as a git-diff. Following the standard used in git diff, the
// path to the "old" version is prefixed with a/, and the "new" with b/:
//
// --- a/.github/workflows/foo.yml
// +++ b/.github/workflows/foo.yml
// @@ -42,13 +42,22 @@
// ...
func getDiff(path, original, patched string) string {
	edits := myers.ComputeEdits(span.URIFromPath(path), original, patched)
	aPath := "a/" + path
	bPath := "b/" + path
	return fmt.Sprint(gotextdiff.ToUnified(aPath, bPath, original, edits))
}
