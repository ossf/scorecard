// Copyright 2021 Security Scorecard Authors
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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/google/go-github/v32/github"
	"github.com/jszwec/csvutil"
	"github.com/shurcooL/githubv4"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"

	"github.com/ossf/scorecard/checks"
	"github.com/ossf/scorecard/cron/bq"
	"github.com/ossf/scorecard/cron/monitoring"
	"github.com/ossf/scorecard/pkg"
	"github.com/ossf/scorecard/repos"
	"github.com/ossf/scorecard/roundtripper"
	"github.com/ossf/scorecard/stats"
)

type Repository struct {
	Repo     string `csv:"repo"`
	Metadata string `csv:"metadata,omitempty"`
}

func startMetricsExporter() (*stackdriver.Exporter, error) {
	exporter, err := monitoring.NewStackDriverExporter()
	if err != nil {
		return nil, fmt.Errorf("error during NewStackDriverExporter: %w", err)
	}
	if err := exporter.StartMetricsExporter(); err != nil {
		return nil, fmt.Errorf("error in StartMetricsExporter: %w", err)
	}

	if err := view.Register(
		&stats.CheckRuntime,
		&stats.OutgoingHTTPRequests); err != nil {
		return nil, fmt.Errorf("error during view.Register: %w", err)
	}
	return exporter, nil
}

func main() {
	bucket := os.Getenv("GCS_BUCKET")
	if bucket == "" {
		log.Fatal("env variable GCS_BUCKET is empty")
	}

	projects, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer projects.Close()

	inputRepos := []Repository{}
	if data, err := ioutil.ReadAll(projects); err != nil {
		panic(err)
	} else if err = csvutil.Unmarshal(data, &inputRepos); err != nil {
		panic(err)
	}

	currTime := time.Now()
	fileName := fmt.Sprintf("%02d-%02d-%d.json",
		currTime.Month(), currTime.Day(), currTime.Year())
	result, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(zap.InfoLevel)
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()
	// Use our custom roundtripper
	rt := roundtripper.NewTransport(ctx, sugar)
	httpClient := &http.Client{
		Transport: rt,
	}
	githubClient := github.NewClient(httpClient)
	graphClient := githubv4.NewClient(httpClient)

	exporter, err := startMetricsExporter()
	if err != nil {
		panic(err)
	}
	defer exporter.Flush()
	defer exporter.StopMetricsExporter()

	for _, r := range inputRepos {
		fmt.Println(r.Repo)

		repoURL := repos.RepoURL{}
		if err := repoURL.Set(r.Repo); err != nil {
			panic(err)
		}
		if err := repoURL.ValidGitHubURL(); err != nil {
			panic(err)
		}

		//nolint
		// FIXME :- deleting branch-protection
		// The branch protection check needs an admin access to the repository.
		// All of the checks from cron would fail and uses another call to the API.
		// This will reduce usage of the API.
		delete(checks.AllChecks, "Branch-Protection")

		repoResult := pkg.RunScorecards(ctx, repoURL, checks.AllChecks, httpClient, githubClient, graphClient)
		repoResult.Date = currTime.Format("2006-01-02")
		if err := repoResult.AsJSON( /*showDetails=*/ true, result); err != nil {
			panic(err)
		}
		//nolint
		logger.Sync() // flushes buffer, if any
	}
	result.Close()

	// copying the file to the GCS bucket
	if err := exec.Command("gsutil", "cp", fileName, fmt.Sprintf("gs://%s", bucket)).Run(); err != nil {
		panic(err)
	}
	//copying the results to the latest.json
	//nolint
	if err := exec.Command("gsutil", "cp", fmt.Sprintf("gs://%s/%s", bucket, fileName),
		fmt.Sprintf("gs://%s/latest.json", bucket)).Run(); err != nil {
		panic(err)
	}

	if startBQDataTransfer, lookup := os.LookupEnv("SCORECARD_START_BQ_TRNSFER"); lookup && startBQDataTransfer != "" {
		if parsedBool, err := strconv.ParseBool(startBQDataTransfer); parsedBool && err == nil {
			// start BQ data transfer job
			if err := bq.StartDataTransferJob(ctx, fmt.Sprintf("gs://%s", bucket), "latest.json"); err != nil {
				panic(err)
			}
		}
	}

	fmt.Println("Finished")
}
