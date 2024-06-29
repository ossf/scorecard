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
TODO:
  - Handle array inputs (i.e. workflow using `github.event.commits[0]` and
    `github.event.commits[1]`, which would duplicate $COMMIT_MESSAGE). Currently throws
    an error on validation.
  - Handle use of synonyms (i.e `commits.*?\.author\.email` and
    `head_commit\.author\.email“, which would duplicate $AUTHOR_EMAIL). Currently throws
    an error on validation.
*/
package patch

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Fixes the script injection identified by the finding and returns a unified diff
// users can apply (with `git apply` or `patch`) to fix the workflow themselves.
// Should an error occur, it is handled and an empty patch is returned.
func GeneratePatch(f checker.File, content string, workflow *actionlint.Workflow, workflowErrs []*actionlint.Error) (string, error) {
	patchedWorkflow, err := patchWorkflow(f, content, workflow)
	if err != nil {
		return "", err
	}
	errs := validatePatchedWorkflow(patchedWorkflow, workflowErrs)
	if len(errs) > 0 {
		return "", fileparser.FormatActionlintError(errs)
	}
	return getDiff(f.Path, content, patchedWorkflow)
}

// Returns a patched version of the workflow without the script injection finding.
func patchWorkflow(f checker.File, content string, workflow *actionlint.Workflow) (string, error) {
	unsafeVar := strings.Trim(f.Snippet, " ")
	runCmdIndex := f.Offset - 1

	lines := strings.Split(string(content), "\n")

	existingEnvvars := parseExistingEnvvars(workflow)
	envvarPatterns := buildUnsafePatterns()
	useExistingEnvvars(envvarPatterns, existingEnvvars, unsafeVar)

	unsafePattern, ok := getUnsafePattern(unsafeVar, envvarPatterns)
	if !ok {
		// TODO: return meaningful error for logging, even if we don't throw it.
		return "", errors.New("AAA")
	}

	lines = replaceUnsafeVarWithEnvvar(lines, unsafePattern, runCmdIndex)

	lines, ok = addEnvvarsToGlobalEnv(lines, existingEnvvars, unsafePattern, unsafeVar)
	if !ok {
		// TODO: return meaningful error for logging, even if we don't throw it.
		return "", errors.New("AAA")
	}

	return strings.Join(lines, "\n"), nil
}

func useExistingEnvvars(unsafePatterns map[string]unsafePattern, existingEnvvars map[string]string, unsafeVar string) {
	if envvar, ok := existingEnvvars[unsafeVar]; ok {
		// There already exists an envvar handling our unsafe variable.
		// Use that envvar instead of creating a separate envvar with the same value.
		pattern, ok := getUnsafePattern(unsafeVar, unsafePatterns)
		if !ok {
			return
		}

		pattern.envvarName = envvar
		unsafePatterns[pattern.ghVarName] = pattern
		return
	}

	// if there's an envvar with the same name as what we'd use, add a "_1" suffix to
	// our envvar name to avoid conflicts. Clumsy but works, and should be rare.
	for _, e := range existingEnvvars {
		for k, p := range unsafePatterns {
			if e == p.envvarName {
				p.envvarName += "_1"
				unsafePatterns[k] = p
			}
		}
	}
}

// Adds a new global environment to a workflow. Assumes a global environment does not
// yet exist.
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

// Identifies the "global" indentation, as defined by the indentation on the required
// `on:` block. Will equal 0 in almost all cases.
func findGlobalIndentation(lines []string) (int, bool) {
	r := regexp.MustCompile(`^\s*on:`)
	for _, line := range lines {
		if r.MatchString(line) {
			return getIndent(line), true
		}
	}

	return -1, false
}

// Detects whether a global `env:` block already exists.
//
// Returns:
//   - int: the index for the line where the `env:` block is declared
//   - int: the indentation used for the declared environment variables
//
// The first two values return -1 if the `env` block doesn't exist
func findExistingEnv(lines []string, globalIndent int) (int, int) {
	num_lines := len(lines)
	indent := strings.Repeat(" ", globalIndent)

	// regex to detect the global `env:` block
	labelRegex := regexp.MustCompile(indent + "env:")

	var currPos int
	var line string
	for currPos, line = range lines {
		if labelRegex.MatchString(line) {
			break
		}
	}

	if currPos >= num_lines-1 {
		// there must be at least one more line
		return -1, -1
	}

	currPos++            // move to line after `env:`
	insertPos := currPos // mark the position where new envvars will be added
	envvarIndent := getIndent(lines[currPos])
	for i, line := range lines[currPos:] {
		if isBlankOrComment(line) {
			continue
		}

		if isParentLevelIndent(line, globalIndent) {
			// no longer declaring envvars
			break
		}

		insertPos = currPos + i + 1
	}

	return insertPos, envvarIndent
}

// Identifies the line where a new `env:` block should be inserted: right above the
// `jobs:` label.
//
// Returns:
//   - int: the index for the line where the `env:` block should be inserted
//   - bool: whether the `jobs:` block was found. Should always be `true`
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
	ghVarName    string
	idRegex      *regexp.Regexp
	replaceRegex *regexp.Regexp
}

func newUnsafePattern(e, p string) unsafePattern {
	return unsafePattern{
		envvarName:   e,
		ghVarName:    p,
		idRegex:      regexp.MustCompile(p),
		replaceRegex: regexp.MustCompile(`{{\s*.*?` + p + `.*?\s*}}`),
	}
}

func getUnsafePattern(unsafeVar string, unsafePatterns map[string]unsafePattern) (unsafePattern, bool) {
	for _, p := range unsafePatterns {
		if p.idRegex.MatchString(unsafeVar) {
			p := p
			return p, true
		}
	}
	return unsafePattern{}, false
}

// Adds the necessary environment variable to the global `env:` block, if it exists.
// If the `env:` block does not exist, it is created right above the `jobs:` label.
func addEnvvarsToGlobalEnv(lines []string, existingEnvvars map[string]string, pattern unsafePattern, unsafeVar string) ([]string, bool) {
	globalIndentation, ok := findGlobalIndentation(lines)
	if !ok {
		// invalid workflow, could not determine global indentation
		return lines, false
	}

	if _, ok := existingEnvvars[unsafeVar]; ok {
		// an existing envvar already handles this unsafe var, we can simply use it
		return lines, true
	}

	var insertPos, envvarIndent int
	if len(existingEnvvars) > 0 {
		insertPos, envvarIndent = findExistingEnv(lines, globalIndentation)
	} else {
		lines, insertPos, ok = addNewGlobalEnv(lines, globalIndentation)
		if !ok {
			return lines, ok
		}

		// position now points to `env:`, insert variables below it
		insertPos += 1
		envvarIndent = globalIndentation + getDefaultIndent(lines)
	}

	envvarDefinition := fmt.Sprintf("%s: ${{ %s }}", pattern.envvarName, unsafeVar)
	lines = slices.Insert(lines, insertPos,
		strings.Repeat(" ", envvarIndent)+envvarDefinition,
	)

	return lines, true
}

func parseExistingEnvvars(workflow *actionlint.Workflow) map[string]string {
	envvars := make(map[string]string)

	if workflow.Env == nil {
		return envvars
	}

	r := regexp.MustCompile(`\$\{\{\s*(github\.[^\s]*?)\s*}}`)
	for _, v := range workflow.Env.Vars {
		value := v.Value.Value

		if strings.Contains(value, "${{") {
			// extract simple variable definition (without brackets, etc)
			m := r.FindStringSubmatch(value)
			if len(m) == 2 {
				value = m[1]
				envvars[value] = v.Name.Value
			} else {
				envvars[v.Value.Value] = v.Name.Value
			}
		} else {
			envvars[v.Value.Value] = v.Name.Value
		}
	}

	return envvars
}

// Replaces all instances of the given script injection variable with the safe
// environment variable.
func replaceUnsafeVarWithEnvvar(lines []string, pattern unsafePattern, runIndex uint) []string {
	runIndent := getIndent(lines[runIndex])
	for i, line := range lines[runIndex:] {
		currLine := int(runIndex) + i
		if i > 0 && isParentLevelIndent(lines[currLine], runIndent) {
			// anything at the same indent as the first line of the  `- run:` block will
			// mean the end of the run block.
			break
		}
		lines[currLine] = pattern.replaceRegex.ReplaceAllString(line, pattern.envvarName)
	}

	return lines
}

func buildUnsafePatterns() map[string]unsafePattern {
	unsafePatterns := []unsafePattern{
		newUnsafePattern("AUTHOR_EMAIL", `github\.event\.commits.*?\.author\.email`),
		newUnsafePattern("AUTHOR_EMAIL", `github\.event\.head_commit\.author\.email`),
		newUnsafePattern("AUTHOR_NAME", `github\.event\.commits.*?\.author\.name`),
		newUnsafePattern("AUTHOR_NAME", `github\.event\.head_commit\.author\.name`),
		newUnsafePattern("COMMENT_BODY", `github\.event\.comment\.body`),
		newUnsafePattern("COMMENT_BODY", `github\.event\.issue_comment\.comment\.body`),
		newUnsafePattern("COMMIT_MESSAGE", `github\.event\.commits.*?\.message`),
		newUnsafePattern("COMMIT_MESSAGE", `github\.event\.head_commit\.message`),
		newUnsafePattern("DISCUSSION_TITLE", `github\.event\.discussion\.title`),
		newUnsafePattern("DISCUSSION_BODY", `github\.event\.discussion\.body`),
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
	m := make(map[string]unsafePattern)

	for _, p := range unsafePatterns {
		p := p
		m[p.ghVarName] = p
	}

	return m
}

// Returns the indentation of the given line. The indentation is all whitespace and
// dashes before a key or value.
func getIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " -"))
}

func isBlankOrComment(line string) bool {
	blank := regexp.MustCompile(`^\s*$`)
	comment := regexp.MustCompile(`^\s*#`)

	return blank.MatchString(line) || comment.MatchString(line)
}

// Returns whether the given line is at the same indentation level as the parent scope.
// For example, when walking through the document, parsing `job_foo`:
//
//	job_foo:
//	  runs-on: ubuntu-latest  # looping over these lines, we have
//	  uses: ./actions/foo     # parent_indent = 2 (job_foo's indentation)
//	  ...                     # we know these lines belong to job_foo because
//	  ...                     # they all have indent = 4
//	job_bar:  # this line has job_foo's indentation, so we know job_foo is done
//
// Blank lines and those containing only comments are ignored and always return False.
func isParentLevelIndent(line string, parentIndent int) bool {
	if isBlankOrComment(line) {
		return false
	}
	return getIndent(line) <= parentIndent
}

func getDefaultIndent(lines []string) int {
	jobs := regexp.MustCompile(`^\s*jobs:`)
	var jobsIndex, jobsIndent int
	for i, line := range lines {
		if jobs.MatchString(line) {
			jobsIndex = i
			jobsIndent = getIndent(line)
			break
		}
	}

	jobIndent := jobsIndent + 2 // default value, should never be used
	for _, line := range lines[jobsIndex+1:] {
		if isBlankOrComment(line) {
			continue
		}
		jobIndent = getIndent(line)
		break
	}

	return jobIndent - jobsIndent
}

func validatePatchedWorkflow(content string, originalErrs []*actionlint.Error) []*actionlint.Error {
	_, patchedErrs := actionlint.Parse([]byte(content))
	if len(patchedErrs) == 0 {
		return []*actionlint.Error{}
	}
	if len(originalErrs) == 0 {
		return patchedErrs
	}

	normalizeMsg := func(msg string) string {
		// one of the error messages contains line metadata that may legitimately change
		// after a patch. Only looking at the errors' first sentence eliminates this.
		return strings.Split(msg, ".")[0]
	}

	var newErrs []*actionlint.Error

	o := 0
	orig := originalErrs[o]
	origMsg := normalizeMsg(orig.Message)

	for _, patched := range patchedErrs {
		if o == len(originalErrs) {
			// no more errors in the original workflow, must be an error from our patch
			newErrs = append(newErrs, patched)
			continue
		}

		msg := normalizeMsg(patched.Message)
		if orig.Column == patched.Column && orig.Kind == patched.Kind && origMsg == msg {
			// Matched error, therefore not due to our patch.
			o++
			if o < len(originalErrs) {
				orig = originalErrs[o]
				origMsg = normalizeMsg(orig.Message)
			}
		} else {
			newErrs = append(newErrs, patched)
		}
	}

	return newErrs
}

// Returns the changes between the original and patched workflows as a unified diff
// (the same generated by `git diff` or `diff -u`).
func getDiff(path, original, patched string) (string, error) {
	// initialize an in-memory repository
	repo, err := newInMemoryRepo()
	if err != nil {
		return "", err
	}

	// commit original workflow to in-memory repository
	originalCommit, err := commitWorkflow(path, original, repo)
	if err != nil {
		return "", err
	}

	// commit patched workflow to in-memory repository
	patchedCommit, err := commitWorkflow(path, patched, repo)
	if err != nil {
		return "", err
	}

	return toUnifiedDiff(originalCommit, patchedCommit)
}

// Initializes an in-memory repository
func newInMemoryRepo() (*git.Repository, error) {
	// initialize an in-memory repository
	filesystem := memfs.New()
	repo, err := git.Init(memory.NewStorage(), filesystem)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// Commits the workflow at the given path to the in-memory repository
func commitWorkflow(path, contents string, repo *git.Repository) (*object.Commit, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	filesystem := worktree.Filesystem

	// create (or overwrite) file
	df, err := filesystem.Create(path)
	if err != nil {
		return nil, err
	}

	df.Write([]byte(contents))
	df.Close()

	// commit file to in-memory repository
	worktree.Add(path)
	hash, err := worktree.Commit("x", &git.CommitOptions{})
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

// Returns a unified diff describing the difference between the given commits
func toUnifiedDiff(originalCommit, patchedCommit *object.Commit) (string, error) {
	patch, err := originalCommit.Patch(patchedCommit)
	if err != nil {
		return "", err
	}
	builder := strings.Builder{}
	patch.Encode(&builder)

	return builder.String(), nil
}
