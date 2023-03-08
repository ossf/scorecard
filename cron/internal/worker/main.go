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

// Package main implements cron worker job.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec

	"go.opencensus.io/stats/view"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	githubstats "github.com/ossf/scorecard/v4/clients/githubrepo/stats"
	"github.com/ossf/scorecard/v4/clients/ossfuzz"
	"github.com/ossf/scorecard/v4/cron/config"
	"github.com/ossf/scorecard/v4/cron/data"
	format "github.com/ossf/scorecard/v4/cron/internal/format"
	"github.com/ossf/scorecard/v4/cron/monitoring"
	"github.com/ossf/scorecard/v4/cron/worker"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
	"github.com/ossf/scorecard/v4/policy"
	"github.com/ossf/scorecard/v4/stats"
)

const (
	resultsFile    = "results.json"
	rawResultsFile = "raw.json"
)

var ignoreRuntimeErrors = flag.Bool("ignoreRuntimeErrors", false, "if set to true any runtime errors will be ignored")

type ScorecardWorker struct {
	ctx               context.Context
	logger            *log.Logger
	checkDocs         docs.Doc
	exporter          monitoring.Exporter
	repoClient        clients.RepoClient
	ciiClient         clients.CIIBestPracticesClient
	ossFuzzRepoClient clients.RepoClient
	vulnsClient       clients.VulnerabilitiesClient
	apiBucketURL      string
	rawBucketURL      string
	blacklistedChecks []string
}

func newScorecardWorker() (*ScorecardWorker, error) {
	var err error
	sw := &ScorecardWorker{}
	if sw.checkDocs, err = docs.Read(); err != nil {
		return nil, fmt.Errorf("docs.Read: %w", err)
	}

	if sw.rawBucketURL, err = config.GetRawResultDataBucketURL(); err != nil {
		return nil, fmt.Errorf("docs.GetRawResultDataBucketURL: %w", err)
	}

	if sw.blacklistedChecks, err = config.GetBlacklistedChecks(); err != nil {
		return nil, fmt.Errorf("config.GetBlacklistedChecks: %w", err)
	}

	var ciiDataBucketURL string
	if ciiDataBucketURL, err = config.GetCIIDataBucketURL(); err != nil {
		return nil, fmt.Errorf("config.GetCIIDataBucketURL: %w", err)
	}

	if sw.apiBucketURL, err = config.GetAPIResultsBucketURL(); err != nil {
		return nil, fmt.Errorf("config.GetAPIResultsBucketURL: %w", err)
	}

	sw.ctx = context.Background()
	sw.logger = log.NewLogger(log.InfoLevel)
	sw.repoClient = githubrepo.CreateGithubRepoClient(sw.ctx, sw.logger)
	sw.ciiClient = clients.BlobCIIBestPracticesClient(ciiDataBucketURL)
	if sw.ossFuzzRepoClient, err = ossfuzz.CreateOSSFuzzClientEager(ossfuzz.StatusURL); err != nil {
		return nil, fmt.Errorf("ossfuzz.CreateOSSFuzzClientEager: %w", err)
	}

	sw.vulnsClient = clients.DefaultVulnerabilitiesClient()

	if sw.exporter, err = startMetricsExporter(); err != nil {
		return nil, fmt.Errorf("startMetricsExporter: %w", err)
	}

	// Exposed for monitoring runtime profiles
	go func() {
		// TODO(log): Previously Fatal. Need to handle the error here.
		//nolint:gosec // not internet facing.
		sw.logger.Info(fmt.Sprintf("%v", http.ListenAndServe(":8080", nil)))
	}()
	return sw, nil
}

func (sw *ScorecardWorker) Close() {
	sw.exporter.StopMetricsExporter()
	sw.ossFuzzRepoClient.Close()
}

func (sw *ScorecardWorker) Process(ctx context.Context, req *data.ScorecardBatchRequest, bucketURL string) error {
	return processRequest(ctx, req, sw.blacklistedChecks, bucketURL, sw.rawBucketURL, sw.apiBucketURL,
		sw.checkDocs, sw.repoClient, sw.ossFuzzRepoClient, sw.ciiClient, sw.vulnsClient, sw.logger)
}

func (sw *ScorecardWorker) PostProcess() {
	sw.exporter.Flush()
}

//nolint:gocognit
func processRequest(ctx context.Context,
	batchRequest *data.ScorecardBatchRequest,
	blacklistedChecks []string, bucketURL, rawBucketURL, apiBucketURL string,
	checkDocs docs.Doc,
	repoClient clients.RepoClient, ossFuzzRepoClient clients.RepoClient,
	ciiClient clients.CIIBestPracticesClient,
	vulnsClient clients.VulnerabilitiesClient,
	logger *log.Logger,
) error {
	filename := worker.ResultFilename(batchRequest)

	var buffer2 bytes.Buffer
	var rawBuffer bytes.Buffer
	// TODO: run Scorecard for each repo in a separate thread.
	for _, repoReq := range batchRequest.GetRepos() {
		logger.Info(fmt.Sprintf("Running Scorecard for repo: %s", *repoReq.Url))
		repo, err := githubrepo.MakeGithubRepo(*repoReq.Url)
		if err != nil {
			// TODO(log): Previously Warn. Consider logging an error here.
			logger.Info(fmt.Sprintf("invalid GitHub URL: %v", err))
			continue
		}
		repo.AppendMetadata(repoReq.Metadata...)

		commitSHA := clients.HeadSHA
		requiredRequestType := []checker.RequestType{}
		if repoReq.Commit != nil && *repoReq.Commit != clients.HeadSHA {
			commitSHA = *repoReq.Commit
			requiredRequestType = append(requiredRequestType, checker.CommitBased)
		}
		checksToRun, err := policy.GetEnabled(nil /*policy*/, nil /*checks*/, requiredRequestType)
		if err != nil {
			return fmt.Errorf("error during policy.GetEnabled: %w", err)
		}
		for _, check := range blacklistedChecks {
			delete(checksToRun, check)
		}

		result, err := pkg.RunScorecard(ctx, repo, commitSHA, 0, checksToRun,
			repoClient, ossFuzzRepoClient, ciiClient, vulnsClient)
		if errors.Is(err, sce.ErrRepoUnreachable) {
			// Not accessible repo - continue.
			continue
		}
		if err != nil {
			return fmt.Errorf("error during RunScorecard: %w", err)
		}
		for checkIndex := range result.Checks {
			check := &result.Checks[checkIndex]
			if !errors.Is(check.Error, sce.ErrScorecardInternal) {
				continue
			}
			errorMsg := fmt.Sprintf("check %s has a runtime error: %v", check.Name, check.Error)
			if !(*ignoreRuntimeErrors) {
				//nolint: goerr113
				return errors.New(errorMsg)
			}
			// TODO(log): Previously Warn. Consider logging an error here.
			logger.Info(errorMsg)
		}
		result.Date = batchRequest.GetJobTime().AsTime()

		if err := format.AsJSON2(&result, true /*showDetails*/, log.InfoLevel, checkDocs, &buffer2); err != nil {
			return fmt.Errorf("error during result.AsJSON2: %w", err)
		}
		// these are for exporting results to GCS for API consumption
		var exportBuffer bytes.Buffer
		var exportRawBuffer bytes.Buffer

		if err := format.AsJSON2(&result, true /*showDetails*/, log.InfoLevel, checkDocs, &exportBuffer); err != nil {
			return fmt.Errorf("error during result.AsJSON2 for export: %w", err)
		}
		if err := format.AsRawJSON(&result, &exportRawBuffer); err != nil {
			return fmt.Errorf("error during result.AsRawJSON for export: %w", err)
		}
		exportPath := fmt.Sprintf("%s/%s", repo.URI(), resultsFile)
		exportCommitSHAPath := fmt.Sprintf("%s/%s/%s", repo.URI(), result.Repo.CommitSHA, resultsFile)
		exportRawPath := fmt.Sprintf("%s/%s", repo.URI(), rawResultsFile)
		exportRawCommitSHAPath := fmt.Sprintf("%s/%s/%s", repo.URI(), result.Repo.CommitSHA, rawResultsFile)

		// Raw result.
		if err := format.AsRawJSON(&result, &rawBuffer); err != nil {
			return fmt.Errorf("error during result.AsRawJSON: %w", err)
		}

		// These are results without the commit SHA which represents the latest commit.
		if err := data.WriteToBlobStore(ctx, apiBucketURL, exportPath, exportBuffer.Bytes()); err != nil {
			return fmt.Errorf("error during writing to exportBucketURL: %w", err)
		}
		// Export result based on commitSHA.
		if err := data.WriteToBlobStore(ctx, apiBucketURL, exportCommitSHAPath, exportBuffer.Bytes()); err != nil {
			return fmt.Errorf("error during exportBucketURL with commit SHA: %w", err)
		}
		// Export raw result.
		if err := data.WriteToBlobStore(ctx, apiBucketURL, exportRawPath, exportRawBuffer.Bytes()); err != nil {
			return fmt.Errorf("error during writing to exportBucketURL for raw results: %w", err)
		}
		if err := data.WriteToBlobStore(ctx, apiBucketURL, exportRawCommitSHAPath, exportRawBuffer.Bytes()); err != nil {
			return fmt.Errorf("error during exportBucketURL for raw results with commit SHA: %w", err)
		}
	}

	// Raw result.
	if err := data.WriteToBlobStore(ctx, rawBucketURL, filename, rawBuffer.Bytes()); err != nil {
		return fmt.Errorf("error during WriteToBlobStore2: %w", err)
	}

	// write to the canonical bucket last, as the presence of filename indicates the job was completed.
	// see worker package for details.
	if err := data.WriteToBlobStore(ctx, bucketURL, filename, buffer2.Bytes()); err != nil {
		return fmt.Errorf("error during WriteToBlobStore2: %w", err)
	}

	logger.Info(fmt.Sprintf("Write to shard file successful: %s", filename))

	return nil
}

func startMetricsExporter() (monitoring.Exporter, error) {
	exporter, err := monitoring.GetExporter()
	if err != nil {
		return nil, fmt.Errorf("error during monitoring.GetExporter: %w", err)
	}
	if err := exporter.StartMetricsExporter(); err != nil {
		return nil, fmt.Errorf("error in StartMetricsExporter: %w", err)
	}

	if err := view.Register(
		&stats.CheckRuntime,
		&stats.CheckErrorCount,
		&stats.OutgoingHTTPRequests,
		&githubstats.GithubTokens); err != nil {
		return nil, fmt.Errorf("error during view.Register: %w", err)
	}
	return exporter, nil
}

func main() {
	flag.Parse()
	sw, err := newScorecardWorker()
	if err != nil {
		panic(err)
	}
	defer sw.Close()
	wl := worker.NewWorkLoop(sw)
	if err := wl.Run(); err != nil {
		panic(err)
	}
}
