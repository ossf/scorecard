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

// Package scorecard defines functions for running Scorecard checks on a Repo.
package scorecard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/azuredevopsrepo"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	"github.com/ossf/scorecard/v5/clients/ossfuzz"
	"github.com/ossf/scorecard/v5/config"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/packageclient"
	proberegistration "github.com/ossf/scorecard/v5/internal/probes"
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/policy"
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
	projectClient packageclient.ProjectPackageClient,
) (Result, error) {
	if err := repoClient.InitRepo(repo, commitSHA, commitDepth); err != nil {
		// No need to call sce.WithMessage() since InitRepo will do that for us.
		//nolint:wrapcheck
		return Result{}, err
	}
	defer repoClient.Close()

	versionInfo := version.GetVersionInfo()
	ret := Result{
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
		return Result{}, err
	}
	ret.Repo.CommitSHA = commitSHA

	defaultBranch, err := repoClient.GetDefaultBranchName()
	if err != nil {
		if !errors.Is(err, clients.ErrUnsupportedFeature) {
			return Result{},
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetDefaultBranchName:%v", err.Error()))
		}
		defaultBranch = "unknown"
	}

	resultsCh := make(chan checker.CheckResult)

	localPath, err := repoClient.LocalPath()
	if err != nil {
		return Result{}, fmt.Errorf("RepoClient.LocalPath: %w", err)
	}

	// Set metadata for all checks to use. This is necessary
	// to create remediations from the probe yaml files.
	ret.RawResults.Metadata.Metadata = map[string]string{
		"repository.host":          repo.Host(),
		"repository.name":          strings.TrimPrefix(repo.URI(), repo.Host()+"/"),
		"repository.uri":           repo.URI(),
		"repository.sha1":          commitSHA,
		"repository.defaultBranch": defaultBranch,
		"localPath":                localPath,
	}

	request := &checker.CheckRequest{
		Ctx:                   ctx,
		RepoClient:            repoClient,
		OssFuzzRepo:           ossFuzzRepoClient,
		CIIClient:             ciiClient,
		VulnerabilitiesClient: vulnsClient,
		ProjectClient:         projectClient,
		Repo:                  repo,
		RawResults:            &ret.RawResults,
	}

	// If the user runs probes
	if len(probesToRun) > 0 {
		err = runEnabledProbes(request, probesToRun, &ret)
		if err != nil {
			return Result{}, err
		}
		return ret, nil
	}

	// If the user runs checks
	go runEnabledChecks(ctx, repo, request, checksToRun, resultsCh)

	// get the repository's config file to read annotations
	r, path := findConfigFile(repoClient)
	logger := sclog.NewLogger(sclog.DefaultLevel)

	if r != nil {
		defer r.Close()
		logger.Info(fmt.Sprintf("using maintainer annotations: %s", path))
		c, err := config.Parse(r)
		if err != nil {
			logger.Info(fmt.Sprintf("couldn't parse maintainer annotations: %v", err))
		}
		ret.Config = c
	}

	for result := range resultsCh {
		ret.Checks = append(ret.Checks, result)
		ret.Findings = append(ret.Findings, result.Findings...)
	}
	return ret, nil
}

func findConfigFile(rc clients.RepoClient) (io.ReadCloser, string) {
	// Look for a config file. Return first one regardless of validity
	locs := []string{"scorecard.yml", ".scorecard.yml", ".github/scorecard.yml"}

	for i := range locs {
		cfr, err := rc.GetFileReader(locs[i])
		if err != nil {
			continue
		}
		return cfr, locs[i]
	}

	return nil, ""
}

func runEnabledProbes(request *checker.CheckRequest,
	probesToRun []string,
	ret *Result,
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

type runConfig struct {
	client        clients.RepoClient
	vulnClient    clients.VulnerabilitiesClient
	ciiClient     clients.CIIBestPracticesClient
	projectClient packageclient.ProjectPackageClient
	ossfuzzClient clients.RepoClient
	commit        string
	logLevel      sclog.Level
	checks        []string
	probes        []string
	commitDepth   int
	gitMode       bool
}

type Option func(*runConfig) error

// WithLogLevel configures the log level of the analysis.
func WithLogLevel(level sclog.Level) Option {
	return func(c *runConfig) error {
		c.logLevel = level
		return nil
	}
}

// WithCommitDepth configures the number of commits to analyze.
func WithCommitDepth(depth int) Option {
	return func(c *runConfig) error {
		c.commitDepth = depth
		return nil
	}
}

// WithCommitSHA specifies the repository commit to analyze.
// If this option is not used, the repository is analyzed at HEAD.
func WithCommitSHA(sha string) Option {
	return func(c *runConfig) error {
		c.commit = sha
		return nil
	}
}

// WithChecks specifies checks which should be run during the analysis
// of a project. If this option is not used, all checks are run.
func WithChecks(checks []string) Option {
	return func(c *runConfig) error {
		c.checks = checks
		return nil
	}
}

// WithProbes specifies individual probes which should be run during the
// analysis of a project.
func WithProbes(probes []string) Option {
	return func(c *runConfig) error {
		c.probes = probes
		return nil
	}
}

// WithRepoClient will set the client used to query a repo host or forge
// about the given project.
func WithRepoClient(client clients.RepoClient) Option {
	return func(c *runConfig) error {
		c.client = client
		return nil
	}
}

// WithOSSFuzzClient will set the client used to query OSS-Fuzz about a project's
// integration with OSS-Fuzz.
func WithOSSFuzzClient(client clients.RepoClient) Option {
	return func(c *runConfig) error {
		c.ossfuzzClient = client
		return nil
	}
}

// WithVulnerabilitiesClient will set the client used to query vulnerabilities
// present in a project.
func WithVulnerabilitiesClient(client clients.VulnerabilitiesClient) Option {
	return func(c *runConfig) error {
		c.vulnClient = client
		return nil
	}
}

// WithOpenSSFBestPraticesClient will set the client used to query the OpenSSF
// Best Practice API for data about a project.
func WithOpenSSFBestPraticesClient(client clients.CIIBestPracticesClient) Option {
	return func(c *runConfig) error {
		c.ciiClient = client
		return nil
	}
}

// WithFileModeGit will configure supporting repository clients to download files
// using git. This is useful for repositories which "export-ignore" files in its
// .gitattributes file.
//
// Repository analysis may be slower.
func WithFileModeGit() Option {
	return func(c *runConfig) error {
		c.gitMode = true
		return nil
	}
}

// Run analyzes a given repository and returns the result. You can modify the
// run behavior by passing in [Option] arguments. In the absence of a particular
// option a default is used. Refer to the various Options for details.
func Run(ctx context.Context, repo clients.Repo, opts ...Option) (Result, error) {
	c := runConfig{
		commit:   clients.HeadSHA,
		logLevel: sclog.DefaultLevel,
	}
	for _, option := range opts {
		if err := option(&c); err != nil {
			return Result{}, err
		}
	}
	logger := sclog.NewLogger(c.logLevel)
	if c.ciiClient == nil {
		c.ciiClient = clients.DefaultCIIBestPracticesClient()
	}
	if c.ossfuzzClient == nil {
		c.ossfuzzClient = ossfuzz.CreateOSSFuzzClient(ossfuzz.StatusURL)
	}
	if c.vulnClient == nil {
		c.vulnClient = clients.DefaultVulnerabilitiesClient()
	}
	if c.projectClient == nil {
		c.projectClient = packageclient.CreateDepsDevClient()
	}

	var requiredRequestTypes []checker.RequestType
	var err error
	switch repo.(type) {
	case *localdir.Repo:
		requiredRequestTypes = append(requiredRequestTypes, checker.FileBased)
		if c.client == nil {
			c.client = localdir.CreateLocalDirClient(ctx, logger)
		}
	case *githubrepo.Repo:
		if c.client == nil {
			var opts []githubrepo.Option
			if c.gitMode {
				opts = append(opts, githubrepo.WithFileModeGit())
			}
			client, err := githubrepo.NewRepoClient(ctx, opts...)
			if err != nil {
				return Result{}, fmt.Errorf("creating github client: %w", err)
			}
			c.client = client
		}
	case *gitlabrepo.Repo:
		if c.client == nil {
			c.client, err = gitlabrepo.CreateGitlabClient(ctx, repo.Host())
			if err != nil {
				return Result{}, fmt.Errorf("creating gitlab client: %w", err)
			}
		}
	case *azuredevopsrepo.Repo:
		if c.client == nil {
			c.client, err = azuredevopsrepo.CreateAzureDevOpsClient(ctx, repo)
			if err != nil {
				return Result{}, fmt.Errorf("creating azure devops client: %w", err)
			}
		}
	}

	if !strings.EqualFold(c.commit, clients.HeadSHA) {
		requiredRequestTypes = append(requiredRequestTypes, checker.CommitBased)
	}

	checksToRun, err := policy.GetEnabled(nil, c.checks, requiredRequestTypes)
	if err != nil {
		return Result{}, fmt.Errorf("getting enabled checks: %w", err)
	}

	return runScorecard(ctx, repo, c.commit, c.commitDepth, checksToRun, c.probes,
		c.client, c.ossfuzzClient, c.ciiClient, c.vulnClient, c.projectClient)
}
