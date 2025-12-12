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

package raw

import (
	"fmt"
	"strings"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

func toCheckerTri(t clients.TriState) checker.TriState {
	switch t {
	case clients.TriTrue:
		return checker.TriTrue
	case clients.TriFalse:
		return checker.TriFalse
	default:
		return checker.TriUnknown
	}
}

func SecretScanning(c *checker.CheckRequest) (checker.SecretScanningData, error) {
	s, err := c.RepoClient.GetSecretScanningSignals()
	if err != nil {
		return checker.SecretScanningData{}, sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("GetSecretScanningSignals: %v", err))
	}

	data := checker.SecretScanningData{
		Platform:                  string(s.Platform),
		GHNativeEnabled:           toCheckerTri(s.GHNativeEnabled),
		GHPushProtectionEnabled:   toCheckerTri(s.GHPushProtectionEnabled),
		GLPipelineSecretDetection: s.GLPipelineSecretDetection,
		GLSecretPushProtection:    s.GLSecretPushProtection,
		GLPushRulesPreventSecrets: s.GLPushRulesPreventSecrets,

		ThirdPartyGitleaks:            s.ThirdPartyGitleaks,
		ThirdPartyGitleaksPaths:       append([]string{}, s.ThirdPartyGitleaksPaths...),
		ThirdPartyTruffleHog:          s.ThirdPartyTruffleHog,
		ThirdPartyTruffleHogPaths:     append([]string{}, s.ThirdPartyTruffleHogPaths...),
		ThirdPartyDetectSecrets:       s.ThirdPartyDetectSecrets,
		ThirdPartyDetectSecretsPaths:  append([]string{}, s.ThirdPartyDetectSecretsPaths...),
		ThirdPartyGitSecrets:          s.ThirdPartyGitSecrets,
		ThirdPartyGitSecretsPaths:     append([]string{}, s.ThirdPartyGitSecretsPaths...),
		ThirdPartyGGShield:            s.ThirdPartyGGShield,
		ThirdPartyGGShieldPaths:       append([]string{}, s.ThirdPartyGGShieldPaths...),
		ThirdPartyShhGit:              s.ThirdPartyShhGit,
		ThirdPartyShhGitPaths:         append([]string{}, s.ThirdPartyShhGitPaths...),
		ThirdPartyRepoSupervisor:      s.ThirdPartyRepoSupervisor,
		ThirdPartyRepoSupervisorPaths: append([]string{}, s.ThirdPartyRepoSupervisorPaths...),

		Evidence: s.Evidence,
	}

	// Collect CI run statistics for third-party tools
	ciStats, err := collectThirdPartyCIStats(c.RepoClient, &data)
	if err != nil {
		// Log but don't fail the check if CI stats collection fails
		c.Dlogger.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("Failed to collect third-party CI stats: %v", err),
		})
	}
	data.ThirdPartyCIInfo = ciStats

	return data, nil
}

// collectThirdPartyCIStats analyzes CI runs for each third-party
// secret scanning tool to determine execution patterns
// (periodic vs commit-based) and run frequency.
func collectThirdPartyCIStats(
	c clients.RepoClient,
	data *checker.SecretScanningData,
) (map[string]*checker.ToolCIStats, error) {
	// Initialize stats for detected tools
	stats := initializeToolStats(data)
	if len(stats) == 0 {
		return stats, nil
	}

	// Get recent commits (up to 100)
	commits, err := c.ListCommits()
	if err != nil {
		return stats, fmt.Errorf("ListCommits: %w", err)
	}

	if len(commits) == 0 {
		return stats, nil
	}

	// Limit to last 100 commits
	if len(commits) > 100 {
		commits = commits[:100]
	}

	// Analyze commits for tool runs
	analyzeCommitsForToolRuns(c, commits, stats)

	return stats, nil
}

// initializeToolStats creates ToolCIStats for each
// detected third-party tool.
func initializeToolStats(
	data *checker.SecretScanningData,
) map[string]*checker.ToolCIStats {
	stats := make(map[string]*checker.ToolCIStats)

	// Periodic tools are expected to run on a schedule (e.g., daily/weekly)
	periodicTools := map[string]bool{
		"shhgit":          true,
		"repo-supervisor": true,
	}

	// Map detected tools to their names
	detectedTools := getDetectedTools(data)

	// Initialize stats for each detected tool
	for toolName := range detectedTools {
		pattern := "commit-based"
		if periodicTools[toolName] {
			pattern = "periodic"
		}
		stats[toolName] = &checker.ToolCIStats{
			ToolName:         toolName,
			ExecutionPattern: pattern,
		}
	}

	return stats
}

// getDetectedTools returns a set of detected third-party
// tools from the data.
func getDetectedTools(data *checker.SecretScanningData) map[string]bool {
	detectedTools := make(map[string]bool)
	if data.ThirdPartyGitleaks {
		detectedTools["gitleaks"] = true
	}
	if data.ThirdPartyTruffleHog {
		detectedTools["trufflehog"] = true
	}
	if data.ThirdPartyDetectSecrets {
		detectedTools["detect-secrets"] = true
	}
	if data.ThirdPartyGitSecrets {
		detectedTools["git-secrets"] = true
	}
	if data.ThirdPartyGGShield {
		detectedTools["ggshield"] = true
	}
	if data.ThirdPartyShhGit {
		detectedTools["shhgit"] = true
	}
	if data.ThirdPartyRepoSupervisor {
		detectedTools["repo-supervisor"] = true
	}
	return detectedTools
}

// analyzeCommitsForToolRuns checks each commit for tool runs
// and updates stats.
func analyzeCommitsForToolRuns(
	c clients.RepoClient,
	commits []clients.Commit,
	stats map[string]*checker.ToolCIStats,
) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// Set total commits analyzed for all tools
	for _, toolStats := range stats {
		toolStats.TotalCommitsAnalyzed = len(commits)
	}

	// Analyze each commit for tool-specific runs
	for i := range commits {
		commit := &commits[i]

		// Skip commits without associated merge requests
		if commit.AssociatedMergeRequest.MergedAt.IsZero() {
			continue
		}

		checkRuns, err := c.ListCheckRunsForRef(commit.AssociatedMergeRequest.HeadSHA)
		if err != nil {
			// Continue on error - just skip this commit
			continue
		}

		updateStatsForCommit(commit, checkRuns, stats, thirtyDaysAgo)
	}
}

// updateStatsForCommit updates tool statistics based on
// check runs for a single commit.
func updateStatsForCommit(
	commit *clients.Commit,
	checkRuns []clients.CheckRun,
	stats map[string]*checker.ToolCIStats,
	thirtyDaysAgo time.Time,
) {
	for toolName, toolStats := range stats {
		if !hasToolInCheckRuns(toolName, checkRuns) {
			continue
		}

		toolStats.CommitsWithToolRun++

		// Check if this run is within last 30 days
		if commit.CommittedDate.After(thirtyDaysAgo) {
			toolStats.HasRecentRuns = true
			if toolStats.LastRunDate == "" ||
				commit.CommittedDate.String() > toolStats.LastRunDate {
				toolStats.LastRunDate = commit.CommittedDate.Format("2006-01-02")
			}
		}
	}
}

// hasToolInCheckRuns returns true if the specific tool
// appears in the check runs.
func hasToolInCheckRuns(toolName string, checkRuns []clients.CheckRun) bool {
	// Map tool names to their possible CI identifiers
	toolKeywords := map[string][]string{
		"gitleaks":        {"gitleaks"},
		"trufflehog":      {"trufflehog", "trufflesecurity"},
		"detect-secrets":  {"detect-secrets", "detect_secrets"},
		"git-secrets":     {"git-secrets", "git_secrets"},
		"ggshield":        {"ggshield", "gitguardian"},
		"shhgit":          {"shhgit"},
		"repo-supervisor": {"repo-supervisor", "repo_supervisor", "reposupervisor"},
	}

	keywords, ok := toolKeywords[toolName]
	if !ok {
		return false
	}

	for i := range checkRuns {
		appSlug := strings.ToLower(checkRuns[i].App.Slug)
		url := strings.ToLower(checkRuns[i].URL)

		for _, keyword := range keywords {
			if strings.Contains(appSlug, keyword) || strings.Contains(url, keyword) {
				return true
			}
		}
	}

	return false
}
