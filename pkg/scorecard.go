// Copyright 2020 OpenSSF Scorecard Authors
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

// Package pkg defines fns for running Scorecard checks on a Repo.
package pkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/options"
	"github.com/ossf/scorecard/v4/probes"
	"github.com/ossf/scorecard/v4/probes/zrunner"
)

// ErrorEmptyRepository indicates the repository is empty.
var ErrorEmptyRepository = errors.New("repository empty")

func runEnabledChecks(ctx context.Context,
	repo clients.Repo, raw *checker.RawResults, checksToRun checker.CheckNameToFnMap,
	repoClient clients.RepoClient, ossFuzzRepoClient clients.RepoClient, ciiClient clients.CIIBestPracticesClient,
	vulnsClient clients.VulnerabilitiesClient,
	resultsCh chan checker.CheckResult,
) {
	request := checker.CheckRequest{
		Ctx:                   ctx,
		RepoClient:            repoClient,
		OssFuzzRepo:           ossFuzzRepoClient,
		CIIClient:             ciiClient,
		VulnerabilitiesClient: vulnsClient,
		Repo:                  repo,
		RawResults:            raw,
	}
	wg := sync.WaitGroup{}
	for checkName, checkFn := range checksToRun {
		checkName := checkName
		checkFn := checkFn
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner := checker.NewRunner(
				checkName,
				repo.URI(),
				&request,
			)

			resultsCh <- runner.Run(ctx, checkFn)
		}()
	}
	wg.Wait()
	close(resultsCh)
}

func getRepoCommitHash(r clients.RepoClient) (string, error) {
	commits, err := r.ListCommits()
	if err != nil {
		// allow --local repos to still process
		if errors.Is(err, clients.ErrUnsupportedFeature) {
			return "unknown", nil
		}
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListCommits:%v", err.Error()))
	}

	if len(commits) > 0 {
		return commits[0].SHA, nil
	}
	return "", ErrorEmptyRepository
}

// RunScorecard runs enabled Scorecard checks on a Repo.
func RunScorecard(ctx context.Context,
	repo clients.Repo,
	commitSHA string,
	commitDepth int,
	checksToRun checker.CheckNameToFnMap,
	repoClient clients.RepoClient,
	ossFuzzRepoClient clients.RepoClient,
	ciiClient clients.CIIBestPracticesClient,
	vulnsClient clients.VulnerabilitiesClient,
) (ScorecardResult, error) {
	if err := repoClient.InitRepo(repo, commitSHA, commitDepth); err != nil {
		// No need to call sce.WithMessage() since InitRepo will do that for us.
		//nolint:wrapcheck
		return ScorecardResult{}, err
	}
	defer repoClient.Close()

	versionInfo := version.GetVersionInfo()
	ret := ScorecardResult{
		Repo: RepoInfo{
			Name:      repo.URI(),
			CommitSHA: commitSHA,
		},
		Scorecard: ScorecardInfo{
			Version:   versionInfo.GitVersion,
			CommitSHA: versionInfo.GitCommit,
		},
		Date: time.Now(),
	}

	commitSHA, err := getRepoCommitHash(repoClient)

	if errors.Is(err, ErrorEmptyRepository) {
		return ret, nil
	} else if err != nil {
		return ScorecardResult{}, err
	}

	defaultBranch, err := repoClient.GetDefaultBranchName()
	if err != nil {
		if !errors.Is(err, clients.ErrUnsupportedFeature) {
			return ScorecardResult{},
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetDefaultBranchName:%v", err.Error()))
		}
		defaultBranch = "unknown"
	}

	resultsCh := make(chan checker.CheckResult)

	// Set metadata for all checks to use. This is necessary
	// to create remediations from the probe yaml files.
	ret.RawResults.Metadata.Metadata = map[string]string{
		"repository.host":          repo.Host(),
		"repository.name":          strings.TrimPrefix(repo.URI(), repo.Host()+"/"),
		"repository.uri":           repo.URI(),
		"repository.sha1":          commitSHA,
		"repository.defaultBranch": defaultBranch,
	}

	go runEnabledChecks(ctx, repo, &ret.RawResults, checksToRun,
		repoClient, ossFuzzRepoClient,
		ciiClient, vulnsClient, resultsCh)

	for result := range resultsCh {
		ret.Checks = append(ret.Checks, result)
	}

	if value, _ := os.LookupEnv(options.EnvVarScorecardExperimental); value == "1" {
		// Run the probes.
		var findings []finding.Finding
		// TODO(#3049): only run the probes for checks.
		// NOTE: We will need separate functions to support:
		// - `--probes X,Y`
		// - `--check-definitions-file path/to/config.yml
		// NOTE: we discard the returned error because the errors are
		// already cotained in the findings and we want to return the findings
		// to users.
		// See https://github.com/ossf/scorecard/blob/main/probes/zrunner/runner.go#L34-L45.
		// Note: we discard the error because each probe's error is reported within
		// the probe and we don't want the entire scorecard run to fail if a single error
		// is encountered.
		//nolint:errcheck
		findings, _ = zrunner.Run(&ret.RawResults, probes.All)
		ret.Findings = findings
	}
	return ret, nil
}
