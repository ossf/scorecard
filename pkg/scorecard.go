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

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/config"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	proberegistration "github.com/ossf/scorecard/v5/internal/probes"
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
)

// errEmptyRepository indicates the repository is empty.
var errEmptyRepository = errors.New("repository empty")

func runEnabledChecks(ctx context.Context,
	repo clients.Repo,
	request *checker.CheckRequest,
	checksToRun checker.CheckNameToFnMap,
	resultsCh chan<- checker.CheckResult,
) {
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
				request,
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

	if len(commits) == 0 {
		return "", errEmptyRepository
	}
	return commits[0].SHA, nil
}

func runScorecard(ctx context.Context,
	repo clients.Repo,
	commitSHA string,
	commitDepth int,
	checksToRun checker.CheckNameToFnMap,
	probesToRun []string,
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

	if errors.Is(err, errEmptyRepository) {
		return ret, nil
	} else if err != nil {
		return ScorecardResult{}, err
	}
	ret.Repo.CommitSHA = commitSHA

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

	request := &checker.CheckRequest{
		Ctx:                   ctx,
		RepoClient:            repoClient,
		OssFuzzRepo:           ossFuzzRepoClient,
		CIIClient:             ciiClient,
		VulnerabilitiesClient: vulnsClient,
		Repo:                  repo,
		RawResults:            &ret.RawResults,
	}

	// If the user runs probes
	if len(probesToRun) > 0 {
		err = runEnabledProbes(request, probesToRun, &ret)
		if err != nil {
			return ScorecardResult{}, err
		}
		return ret, nil
	}

	// If the user runs checks
	go runEnabledChecks(ctx, repo, request, checksToRun, resultsCh)

	if os.Getenv(options.EnvVarScorecardExperimental) == "1" {
		// Get configuration
		rc, err := repoClient.GetFileReader("scorecard.yml")
		// If configuration file exists, continue. Otherwise, ignore
		if err == nil {
			defer rc.Close()
			checks := []string{}
			for check := range checksToRun {
				checks = append(checks, check)
			}
			c, err := config.Parse(rc, checks)
			if err != nil {
				logger := sclog.NewLogger(sclog.DefaultLevel)
				logger.Error(err, "parsing configuration file")
			}
			ret.Config = c
		}
	}

	for result := range resultsCh {
		ret.Checks = append(ret.Checks, result)
		ret.Findings = append(ret.Findings, result.Findings...)
	}
	return ret, nil
}

func runEnabledProbes(request *checker.CheckRequest,
	probesToRun []string,
	ret *ScorecardResult,
) error {
	// Add RawResults to request
	err := populateRawResults(request, probesToRun, ret)
	if err != nil {
		return err
	}

	probeFindings := make([]finding.Finding, 0)
	for _, probeName := range probesToRun {
		probe, err := proberegistration.Get(probeName)
		if err != nil {
			return fmt.Errorf("getting probe %q: %w", probeName, err)
		}
		// Run probe
		var findings []finding.Finding
		if probe.IndependentImplementation != nil {
			findings, _, err = probe.IndependentImplementation(request)
		} else {
			findings, _, err = probe.Implementation(&ret.RawResults)
		}
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, "ending run")
		}
		probeFindings = append(probeFindings, findings...)
	}
	ret.Findings = probeFindings
	return nil
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
	return runScorecard(ctx,
		repo,
		commitSHA,
		commitDepth,
		checksToRun,
		[]string{},
		repoClient,
		ossFuzzRepoClient,
		ciiClient,
		vulnsClient,
	)
}

// ExperimentalRunProbes is experimental. Do not depend on it, it may be removed at any point.
func ExperimentalRunProbes(ctx context.Context,
	repo clients.Repo,
	commitSHA string,
	commitDepth int,
	checksToRun checker.CheckNameToFnMap,
	probesToRun []string,
	repoClient clients.RepoClient,
	ossFuzzRepoClient clients.RepoClient,
	ciiClient clients.CIIBestPracticesClient,
	vulnsClient clients.VulnerabilitiesClient,
) (ScorecardResult, error) {
	return runScorecard(ctx,
		repo,
		commitSHA,
		commitDepth,
		checksToRun,
		probesToRun,
		repoClient,
		ossFuzzRepoClient,
		ciiClient,
		vulnsClient,
	)
}
