// Copyright 2023 OpenSSF Scorecard Authors
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
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
)

const CheckSAST = "SAST"

var errInvalid = errors.New("invalid")

var sastTools = map[string]bool{
	"github-advanced-security": true,
	"github-code-scanning":     true,
	"lgtm-com":                 true,
	"sonarcloud":               true,
	"sonarqubecloud":           true,
}

// Common SAST tool name patterns to detect in CheckRun/Status names
// This enables detection of SAST tools running in any CI system (Prow, Jenkins, etc.)
var sastToolPatterns = []string{
	"codeql",
	"sonar",
	"snyk",
	"semgrep",
	"gosec",
	"staticcheck",
	"golangci",
	"pylint",
	"eslint",
	"tslint",
	"bandit",
	"brakeman",
	"flawfinder",
	"shellcheck",
	"trivy",
	"grype",
	"checkov",
	"tfsec",
	"kubesec",
	"hadolint",
	"safety",
	"osv-scanner",
	"govulncheck",
	"bearer",
	"horusec",
	"fortify",
	"checkmarx",
	"veracode",
	"coverity",
	"insider",
	"security-scan",
	"static-analysis",
	"sast",
	"code-scanning",
	"code-analysis",
}

// sastToolExecutionSignatures maps tool names to their specific execution signatures.
// These are exact patterns that indicate the tool is actually running (not just mentioned).
// This significantly reduces false positives from test data, documentation, or logs.
var sastToolExecutionSignatures = map[string][]string{
	"codeql": {
		"codeql database create",
		"codeql database analyze",
		"running codeql",
		"+ codeql",
		"codeql/analyze",
	},
	"sonar": {
		"sonar-scanner",
		"mvn sonar:sonar",
		"gradle sonar",
		"sonarqube scanner",
		"running sonar",
	},
	"snyk": {
		"snyk test",
		"snyk monitor",
		"snyk container test",
		"snyk iac test",
		"snyk code test",
		"running snyk",
		"+ snyk",
	},
	"semgrep": {
		"semgrep scan",
		"semgrep ci",
		"semgrep --config",
		"running semgrep",
		"+ semgrep",
	},
	"gosec": {
		"gosec ./...",
		"gosec -fmt",
		"running gosec",
		"+ gosec",
		"go run github.com/securego/gosec",
	},
	"staticcheck": {
		"staticcheck ./...",
		"running staticcheck",
		"+ staticcheck",
	},
	"golangci": {
		"golangci-lint run",
		"running golangci-lint",
		"+ golangci-lint",
		"golangci/golangci-lint",
	},
	"pylint": {
		"pylint ",
		"running pylint",
		"+ pylint",
		"python -m pylint",
	},
	"eslint": {
		"eslint .",
		"eslint src",
		"npm run eslint",
		"running eslint",
		"+ eslint",
	},
	"bandit": {
		"bandit -r",
		"running bandit",
		"+ bandit",
		"python -m bandit",
	},
	"brakeman": {
		"brakeman -",
		"bundle exec brakeman",
		"running brakeman",
		"+ brakeman",
	},
	"shellcheck": {
		"shellcheck ",
		"running shellcheck",
		"+ shellcheck",
	},
	"trivy": {
		"trivy scan",
		"trivy fs",
		"trivy image",
		"trivy config",
		"running trivy",
		"+ trivy",
	},
	"hadolint": {
		"hadolint ",
		"running hadolint",
		"+ hadolint",
	},
	"checkov": {
		"checkov -d",
		"checkov --directory",
		"running checkov",
		"+ checkov",
	},
	"tfsec": {
		"tfsec .",
		"running tfsec",
		"+ tfsec",
	},
	"grype": {
		"grype scan",
		"grype dir:",
		"running grype",
		"+ grype",
	},
	"osv-scanner": {
		"osv-scanner ",
		"running osv-scanner",
		"+ osv-scanner",
	},
	"govulncheck": {
		"govulncheck ./...",
		"running govulncheck",
		"+ govulncheck",
		"go run golang.org/x/vuln/cmd/govulncheck",
	},
	"bearer": {
		"bearer scan",
		"running bearer",
		"+ bearer",
		// Explicitly NOT: "authorization: bearer" or "bearer token"
	},
	"horusec": {
		"horusec start",
		"running horusec",
		"+ horusec",
	},
}

var allowedConclusions = map[string]bool{"success": true, "neutral": true}

// SAST checks for presence of static analysis tools.
func SAST(c *checker.CheckRequest) (checker.SASTData, error) {
	var data checker.SASTData

	if c.Dlogger != nil {
		c.Dlogger.Debug(&checker.LogMessage{
			Text: "Starting SAST check - looking for static analysis tools...",
		})
	}

	commits, err := sastToolInCheckRuns(c)
	if err != nil {
		return data, err
	}
	data.Commits = commits

	codeQLWorkflows, err := getSastUsesWorkflows(c, "^github/codeql-action/analyze$", checker.CodeQLWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, codeQLWorkflows...)

	sonarWorkflows, err := getSonarWorkflows(c)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, sonarWorkflows...)

	snykWorkflows, err := getSastUsesWorkflows(c, "^snyk/actions/.*", checker.SnykWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, snykWorkflows...)

	pysaWorkflows, err := getSastUsesWorkflows(c, "^facebook/pysa-action$", checker.PysaWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, pysaWorkflows...)

	qodanaWorkflows, err := getSastUsesWorkflows(c, "^JetBrains/qodana-action$", checker.QodanaWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, qodanaWorkflows...)

	hadolintWorkflows, err := getSastUsesWorkflows(c, "^hadolint/hadolint-action$", checker.HadolintWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, hadolintWorkflows...)

	// Check Prow config files for SAST tools
	prowJobs, err := getProwSASTJobs(c)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, prowJobs...)

	if c.Dlogger != nil {
		c.Dlogger.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("SAST check complete: found %d workflow(s), analyzed %d commit(s)",
				len(data.Workflows), len(data.Commits)),
		})
	}

	return data, nil
}

// isSASTToolName checks if the given string contains a known SAST tool name pattern.
// This enables detection of SAST tools in CI systems like Prow, Jenkins, etc.
func isSASTToolName(s string) bool {
	lower := strings.ToLower(s)
	for _, pattern := range sastToolPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// findMatchingSASTPattern returns the matched pattern if found.
func findMatchingSASTPattern(s string) string {
	lower := strings.ToLower(s)
	for _, pattern := range sastToolPatterns {
		if strings.Contains(lower, pattern) {
			return pattern
		}
	}
	return ""
}

const (
	// Maximum size to fetch from Prow logs (1MB should be enough to detect tool signatures).
	maxProwLogBytes = 1 * 1024 * 1024
	// HTTP timeout for fetching Prow logs.
	prowLogTimeout = 10 * time.Second
)

var errProwLogHTTP = errors.New("HTTP error fetching Prow log")

// isProwJobURL checks if a URL points to an actual Prow job (with logs).
// This filters out Prow status contexts like /tide, /pr-history that don't have logs.
func isProwJobURL(url string) bool {
	lower := strings.ToLower(url)

	// Must contain prow-related patterns
	hasProw := strings.Contains(lower, "prow.") ||
		strings.Contains(lower, "/gs/") ||
		strings.Contains(lower, "storage.googleapis.com")

	if !hasProw {
		return false
	}

	// Exclude known non-job Prow URLs that don't have logs
	excludePatterns := []string{
		"/tide",         // Prow's merge automation (status only)
		"/pr-history",   // PR status history page
		"/command-help", // Help pages
		"/plugin-help",  // Plugin documentation
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(lower, pattern) {
			return false
		}
	}

	// Include URLs that look like actual job runs
	// These typically have: /view/gs/, /pr-logs/, /logs/, or job names
	includePatterns := []string{
		"/view/gs/",
		"/pr-logs/",
		"/logs/",
		"storage.googleapis.com",
	}

	for _, pattern := range includePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// If it has "prow." but doesn't match include patterns, it's likely not a job
	return false
}

// convertToLogURL attempts to convert a Prow view URL to direct log file URLs.
// Returns a list of potential log file URLs to try.
func convertToLogURL(prowURL string) []string {
	var urls []string

	// If it's already a storage URL, try appending log files
	if strings.Contains(prowURL, "storage.googleapis.com") {
		urls = append(urls, prowURL+"/build-log.txt", prowURL+"/finished.json")
		return urls
	}

	// Convert prow.k8s.io/view/gs/... to storage.googleapis.com/...
	if strings.Contains(prowURL, "/view/gs/") {
		parts := strings.Split(prowURL, "/view/gs/")
		if len(parts) == 2 {
			storageURL := "https://storage.googleapis.com/" + parts[1]
			urls = append(urls, storageURL+"/build-log.txt", storageURL+"/finished.json")
		}
	}

	// If no conversion worked, try the original URL with log file paths
	if len(urls) == 0 {
		urls = append(urls, prowURL+"/build-log.txt")
	}

	return urls
}

// fetchProwLog attempts to fetch and read a Prow log file with size limits and timeout.
func fetchProwLog(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, prowLogTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	client := &http.Client{
		Timeout: prowLogTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching log: %w", err)
	}
	defer resp.Body.Close()

	// Only accept 200 OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", errProwLogHTTP, resp.StatusCode)
	}

	// Read with size limit
	limitedReader := io.LimitReader(resp.Body, maxProwLogBytes)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("reading log: %w", err)
	}

	return content, nil
}

// isSASTToolExecution checks if a line indicates actual SAST tool execution.
// Uses tool-specific execution signatures for high accuracy.
func isSASTToolExecution(line, tool string) bool {
	lower := strings.ToLower(line)
	lowerTool := strings.ToLower(tool)

	// Skip lines that are clearly not executions
	// 1. Test data (Authorization: Bearer, token examples, etc.)
	if strings.Contains(lower, "authorization:") ||
		strings.Contains(lower, "bearer token") ||
		strings.Contains(lower, "example") ||
		strings.Contains(lower, "test-token") {
		return false
	}

	// 2. Error messages
	if strings.Contains(lower, "failed to") ||
		strings.Contains(lower, "not found") ||
		strings.Contains(lower, "error:") ||
		strings.Contains(lower, "could not find") ||
		strings.Contains(lower, "unable to") {
		return false
	}

	// 3. Use tool-specific signatures if available
	if signatures, ok := sastToolExecutionSignatures[lowerTool]; ok {
		for _, sig := range signatures {
			if strings.Contains(lower, strings.ToLower(sig)) {
				return true
			}
		}
		return false
	}

	// 4. Fallback: generic execution patterns for tools without specific signatures
	executionPatterns := []string{
		"+" + lowerTool,        // + gosec (shell trace)
		"running " + lowerTool, // Running gosec
		lowerTool + " run",     // gosec run
		lowerTool + " scan",    // trivy scan
		lowerTool + " --",      // Tool with flags
	}

	for _, pattern := range executionPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// scanLogForSASTTools scans log content for SAST tool execution patterns.
func scanLogForSASTTools(content []byte, prowURL string, c *checker.CheckRequest) bool {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		for _, tool := range sastToolPatterns {
			if strings.Contains(strings.ToLower(line), tool) {
				if isSASTToolExecution(line, tool) {
					if c.Dlogger != nil {
						c.Dlogger.Debug(&checker.LogMessage{
							Path: prowURL,
							Type: finding.FileTypeURL,
							Text: fmt.Sprintf("[Prow Logs] SAST tool '%s' detected in job output", tool),
						})
					}
					return true
				}
			}
		}
	}
	return false
}

// scanLogForSecurityLinting scans log content for security linting patterns.
func scanLogForSecurityLinting(content []byte, prowURL string, c *checker.CheckRequest) bool {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		if (strings.Contains(line, "running") || strings.Contains(line, "+") || strings.Contains(line, "executing")) &&
			strings.Contains(line, "security") && strings.Contains(line, "lint") {
			if c.Dlogger != nil {
				c.Dlogger.Debug(&checker.LogMessage{
					Path: prowURL,
					Type: finding.FileTypeURL,
					Text: "[Prow Logs] Security linting detected in job output",
				})
			}
			return true
		}
	}
	return false
}

// hasSASTInLogs fetches and scans Prow log content for SAST tool signatures.
// Uses execution pattern matching to reduce false positives.
func hasSASTInLogs(c *checker.CheckRequest, prowURL string) bool {
	logURLs := convertToLogURL(prowURL)

	for _, logURL := range logURLs {
		content, err := fetchProwLog(c.Ctx, logURL)
		if err != nil {
			if c.Dlogger != nil {
				c.Dlogger.Debug(&checker.LogMessage{
					Text: fmt.Sprintf("[Prow Logs] Could not fetch %s: %v", logURL, err),
				})
			}
			continue
		}

		if scanLogForSASTTools(content, prowURL, c) {
			return true
		}

		if scanLogForSecurityLinting(content, prowURL, c) {
			return true
		}
	}

	return false
}

// checkStatusesForSAST checks commit statuses for SAST tool execution via pattern matching.
// Only logs when a SAST tool is found to reduce noise.
// Fetches Prow logs in parallel for improved performance.
func checkStatusesForSAST(c *checker.CheckRequest, statuses []clients.Status) bool {
	// Skip logging if no statuses to check (reduces noise for repos without statuses)
	if len(statuses) == 0 {
		return false
	}

	// First pass: check pattern matching (fast, synchronous)
	for _, status := range statuses {
		if status.State != "success" {
			continue
		}
		// Check if status context or URL contains a SAST tool name
		matchedInContext := findMatchingSASTPattern(status.Context)
		matchedInURL := findMatchingSASTPattern(status.TargetURL)

		if matchedInContext != "" || matchedInURL != "" {
			if c.Dlogger != nil {
				matchInfo := ""
				if matchedInContext != "" {
					matchInfo = fmt.Sprintf("pattern '%s' in context '%s'", matchedInContext, status.Context)
				} else {
					matchInfo = fmt.Sprintf("pattern '%s' in URL", matchedInURL)
				}
				c.Dlogger.Debug(&checker.LogMessage{
					Path: status.TargetURL,
					Type: finding.FileTypeURL,
					Text: fmt.Sprintf("[CI Status Check] SAST tool detected: %s", matchInfo),
				})
			}
			return true
		}
	}

	// Second pass: collect Prow URLs for parallel log scanning
	var prowURLs []string
	for _, status := range statuses {
		if status.State == "success" && isProwJobURL(status.TargetURL) {
			prowURLs = append(prowURLs, status.TargetURL)
		}
	}

	if len(prowURLs) == 0 {
		return false
	}

	// Parallel log scanning using goroutines
	type result struct {
		url   string
		found bool
	}
	resultChan := make(chan result, len(prowURLs))
	var wg sync.WaitGroup

	for _, url := range prowURLs {
		wg.Add(1)
		go func(prowURL string) {
			defer wg.Done()
			found := hasSASTInLogs(c, prowURL)
			resultChan <- result{url: prowURL, found: found}
		}(url)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for res := range resultChan {
		if res.found {
			return true
		}
	}

	// Only log once if we checked statuses but found nothing
	// (This still happens per commit, but only when statuses exist)
	return false
}

//nolint:gocognit // TODO: refactor to reduce complexity
func sastToolInCheckRuns(c *checker.CheckRequest) ([]checker.SASTCommit, error) {
	var sastCommits []checker.SASTCommit
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		// ignoring check for local dir
		if errors.Is(err, clients.ErrUnsupportedFeature) {
			return sastCommits, nil
		}
		return sastCommits,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListCommits: %v", err))
	}

	if c.Dlogger != nil {
		c.Dlogger.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("[CheckRuns] Analyzing %d commit(s) for SAST tools", len(commits)),
		})
	}

	for i := range commits {
		pr := commits[i].AssociatedMergeRequest
		// TODO(#575): We ignore associated PRs if Scorecard is being run on a fork
		// but the PR was created in the original repo.
		if pr.MergedAt.IsZero() {
			continue
		}

		checked := false
		crs, err := c.RepoClient.ListCheckRunsForRef(pr.HeadSHA)
		if err != nil {
			return sastCommits,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Checks.ListCheckRunsForRef: %v", err))
		}
		// Note: crs may be `nil`: in this case
		// the loop below will be skipped.
		for _, cr := range crs {
			if cr.Status != "completed" {
				continue
			}
			if !allowedConclusions[cr.Conclusion] {
				continue
			}
			// Check for known SAST tool app slugs (GitHub-specific)
			if sastTools[cr.App.Slug] {
				if c.Dlogger != nil {
					c.Dlogger.Debug(&checker.LogMessage{
						Path: cr.URL,
						Type: finding.FileTypeURL,
						Text: fmt.Sprintf("tool detected: %v", cr.App.Slug),
					})
				}
				checked = true
				break
			}
			// Check for SAST tool patterns in app slug or URL (works for Prow, Jenkins, etc.)
			if isSASTToolName(cr.App.Slug) || isSASTToolName(cr.URL) {
				if c.Dlogger != nil {
					c.Dlogger.Debug(&checker.LogMessage{
						Path: cr.URL,
						Type: finding.FileTypeURL,
						Text: fmt.Sprintf("SAST tool pattern detected in: %v", cr.App.Slug),
					})
				}
				checked = true
				break
			}
		}

		// If not found in CheckRuns, also check commit Statuses (used by Prow, Jenkins, etc.)
		if !checked {
			statuses, err := c.RepoClient.ListStatuses(pr.HeadSHA)
			if err != nil {
				// Ignore error if statuses are not supported
				if !errors.Is(err, clients.ErrUnsupportedFeature) {
					return sastCommits,
						sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.ListStatuses: %v", err))
				}
			}
			checked = checkStatusesForSAST(c, statuses)
		}
		sastCommit := checker.SASTCommit{
			CommittedDate:          commits[i].CommittedDate,
			Message:                commits[i].Message,
			SHA:                    commits[i].SHA,
			AssociatedMergeRequest: commits[i].AssociatedMergeRequest,
			Committer:              commits[i].Committer,
			Compliant:              checked,
		}
		sastCommits = append(sastCommits, sastCommit)
	}
	return sastCommits, nil
}

// getSastUsesWorkflows matches if the "uses" field of a GitHub action matches
// a given regex by way of usesRegex. Each workflow that matches the usesRegex
// is appended to the slice that is returned.
func getSastUsesWorkflows(
	c *checker.CheckRequest,
	usesRegex string,
	checkerType checker.SASTWorkflowType,
) ([]checker.SASTWorkflow, error) {
	var workflowPaths []string
	var sastWorkflows []checker.SASTWorkflow
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, searchGitHubActionWorkflowUseRegex, &workflowPaths, usesRegex)
	if err != nil {
		return sastWorkflows, err
	}
	for _, path := range workflowPaths {
		sastWorkflow := checker.SASTWorkflow{
			File: checker.File{
				Path:   path,
				Offset: checker.OffsetDefault,
				Type:   finding.FileTypeSource,
			},
			Type: checkerType,
		}

		sastWorkflows = append(sastWorkflows, sastWorkflow)
	}
	return sastWorkflows, nil
}

var searchGitHubActionWorkflowUseRegex fileparser.DoWhileTrueOnFileContent = func(path string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}

	if len(args) != 2 {
		return false, fmt.Errorf(
			"searchGitHubActionWorkflowUseRegex requires exactly 2 arguments: %w", errInvalid)
	}

	// Verify the type of the data.
	paths, ok := args[0].(*[]string)
	if !ok {
		return false, fmt.Errorf(
			"searchGitHubActionWorkflowUseRegex expects arg[0] of type *[]string: %w", errInvalid)
	}

	usesRegex, ok := args[1].(string)
	if !ok {
		return false, fmt.Errorf(
			"searchGitHubActionWorkflowUseRegex expects arg[1] of type string: %w", errInvalid)
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		return false, fileparser.FormatActionlintError(errs)
	}

	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			e, ok := step.Exec.(*actionlint.ExecAction)
			if !ok || e == nil || e.Uses == nil {
				continue
			}
			// Parse out repo / SHA.
			uses := strings.TrimPrefix(e.Uses.Value, "actions://")
			action, _, _ := strings.Cut(uses, "@")
			re := regexp.MustCompile(usesRegex)
			if re.MatchString(action) {
				*paths = append(*paths, path)
			}
		}
	}
	return true, nil
}

type sonarConfig struct {
	url  string
	file checker.File
}

func getSonarWorkflows(c *checker.CheckRequest) ([]checker.SASTWorkflow, error) {
	var config []sonarConfig
	var sastWorkflows []checker.SASTWorkflow
	// in the future, we may want to use ListFiles instead, so we don't open every file
	err := fileparser.OnMatchingFileReaderDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*",
		CaseSensitive: false,
	}, validateSonarConfig, &config)
	if err != nil {
		return sastWorkflows, err
	}
	for _, result := range config {
		sastWorkflow := checker.SASTWorkflow{
			File: checker.File{
				Path:      result.file.Path,
				Offset:    result.file.Offset,
				EndOffset: result.file.EndOffset,
				Type:      result.file.Type,
				Snippet:   result.url,
			},
			Type: checker.SonarWorkflow,
		}

		sastWorkflows = append(sastWorkflows, sastWorkflow)
	}
	return sastWorkflows, nil
}

// Check file content.
var validateSonarConfig fileparser.DoWhileTrueOnFileReader = func(pathfn string,
	reader io.Reader,
	args ...interface{},
) (bool, error) {
	if !strings.EqualFold(path.Base(pathfn), "pom.xml") {
		return true, nil
	}

	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateSonarConfig requires exactly 1 argument: %w", errInvalid)
	}

	// Verify the type of the data.
	pdata, ok := args[0].(*[]sonarConfig)
	if !ok {
		return false, fmt.Errorf(
			"validateSonarConfig expects arg[0] of type *[]sonarConfig]: %w", errInvalid)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return false, fmt.Errorf("read file: %w", err)
	}
	regex := regexp.MustCompile(`<sonar\.host\.url>\s*(\S+)\s*<\/sonar\.host\.url>`)
	match := regex.FindSubmatch(content)

	if len(match) < 2 {
		return true, nil
	}
	offset, err := findLine(content, []byte("<sonar.host.url>"))
	if err != nil {
		return false, err
	}

	endOffset, err := findLine(content, []byte("</sonar.host.url>"))
	if err != nil {
		return false, err
	}

	*pdata = append(*pdata, sonarConfig{
		url: string(match[1]),
		file: checker.File{
			Path:      pathfn,
			Type:      finding.FileTypeSource,
			Offset:    offset,
			EndOffset: endOffset,
		},
	})

	return true, nil
}

func findLine(content, data []byte) (uint, error) {
	r := bytes.NewReader(content)
	scanner := bufio.NewScanner(r)

	var line uint
	for scanner.Scan() {
		line++
		if bytes.Contains(scanner.Bytes(), data) {
			return line, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scanner.Err(): %w", err)
	}

	return 0, nil
}
