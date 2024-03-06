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
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
)

const (
	assumedIndent = 2
)

func parseDiff(diff string) string {
	i := strings.Index(diff, "\"\"\"\n")
	if i == -1 {
		return diff
	}

	diff = diff[i+4:]
	i = strings.LastIndex(diff, "\"\"\"")
	if i == -1 {
		return diff
	}

	return diff[:i]
}

// Placeholder function that should receive the file of a workflow and
// return the end result of the Script Injection patch
//
// TODO: Receive the dangerous workflow as parameter.
func GeneratePatch(f checker.File) string {
	// TODO: Implement
	// example:
	// type scriptInjection
	// path {.github/workflows/active-elastic-job~active-elastic-job~build.yml  github.head_ref  91 0 0 1}
	// snippet github.head_ref

	// for testing, while we figure out how to get the full path
	path := path.Join("/home/pnacht_google_com/temp_test", f.Path)

	blob, err := os.ReadFile(path)

	if err != nil {
		return ""
	}

	lines := strings.Split(string(blob), "\n")

	globalIndentation, ok := findGlobalIndentation(lines)

	if !ok {
		// invalid workflow, could not determine global indentation
		return ""
	}

	envPos, envvarIndent, exists := findExistingEnv(lines, globalIndentation)

	if !exists {
		envPos, ok = findNewEnvPos(lines, globalIndentation)

		if !ok {
			// invalid workflow, could not determine location for new environment
			return ""
		}

		label := strings.Repeat(" ", globalIndentation) + "env:"
		lines = slices.Insert(lines, envPos, []string{label, ""}...)
		envPos += 1 // position now points to `env:`, insert variables below it
		envvarIndent = globalIndentation + assumedIndent
	}

	envvar, err := convertUnsafeVarToEnvvar(f.Snippet)
	if err != nil {
		fmt.Printf("%v", err)
	}
	lines = slices.Insert(lines, envPos, strings.Repeat(" ", envvarIndent)+envvar)

	for _, line := range lines {
		fmt.Println(line)
	}

	src := `asasas
hello """ola"""
	message=$(echo "${{ github.event.head_commit.message }}" | tail -n +3)
adios`
	dst := `asasas
hello """ola"""
	message=$(echo $COMMIT | tail -n +3)
adios`
	return parseDiff(cmp.Diff(src, dst))
}

func findGlobalIndentation(lines []string) (int, bool) {
	r := regexp.MustCompile(`^\W*on:`)
	for _, line := range lines {
		if r.MatchString(line) {
			return len(line) - len(strings.TrimLeft(line, " ")), true
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
	envvarIndent := len(lines[i]) - len(strings.TrimLeft(lines[i], " "))
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

var unsafeVarToEnvvar = map[*regexp.Regexp]string{
	regexp.MustCompile(`issue\.title`):                             "ISSUE_TITLE",
	regexp.MustCompile(`issue\.body`):                              "ISSUE_BODY",
	regexp.MustCompile(`pull_request\.title`):                      "PR_TITLE",
	regexp.MustCompile(`pull_request\.body`):                       "PR_BODY",
	regexp.MustCompile(`comment\.body`):                            "COMMENT_BODY",
	regexp.MustCompile(`review\.body`):                             "REVIEW_BODY",
	regexp.MustCompile(`review_comment\.body`):                     "REVIEW_COMMENT_BODY",
	regexp.MustCompile(`pages.*\.page_name`):                       "PAGE_NAME",
	regexp.MustCompile(`commits.*\.message`):                       "COMMIT_MESSAGE",
	regexp.MustCompile(`head_commit\.message`):                     "COMMIT_MESSAGE",
	regexp.MustCompile(`head_commit\.author\.email`):               "AUTHOR_EMAIL",
	regexp.MustCompile(`head_commit\.author\.name`):                "AUTHOR_NAME",
	regexp.MustCompile(`commits.*\.author\.email`):                 "AUTHOR_EMAIL",
	regexp.MustCompile(`commits.*\.author\.name`):                  "AUTHOR_NAME",
	regexp.MustCompile(`pull_request\.head\.ref`):                  "PR_HEAD_REF",
	regexp.MustCompile(`pull_request\.head\.label`):                "PR_HEAD_LABEL",
	regexp.MustCompile(`pull_request\.head\.repo\.default_branch`): "PR_DEFAULT_BRANCH",
	regexp.MustCompile(`github\.head_ref`):                         "HEAD_REF",
}

func convertUnsafeVarToEnvvar(unsafeVar string) (string, error) {
	for regex, envvar := range unsafeVarToEnvvar {
		if regex.MatchString(unsafeVar) {
			return fmt.Sprintf("%s: %s", envvar, unsafeVar), nil
		}
	}
	return "", sce.WithMessage(sce.ErrScorecardInternal,
		fmt.Sprintf(
			"Detected unsafe variable '%s', but could not find a compatible envvar name",
			unsafeVar))
}

func replaceUnsafeVarWithEnvvar(line string, unsafeVar string, envvar string) string {
	r := regexp.MustCompile(`${{\W*` + unsafeVar + `\W*}}`)
	return r.ReplaceAllString(line, "$"+envvar)
}
