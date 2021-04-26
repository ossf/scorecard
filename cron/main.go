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
	"os"
	"os/exec"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/ossf/scorecard/checks"
	"github.com/ossf/scorecard/pkg"
	"github.com/ossf/scorecard/repos"
	"go.uber.org/zap"
)

type Repository struct {
	Repo     string `csv:"repo"`
	Metadata string `csv:"metadata,omitempty"`
}

func main() {
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

	fileName := fmt.Sprintf("%02d-%02d-%d.json",
		time.Now().Month(), time.Now().Day(), time.Now().Year())
	result, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	for _, r := range inputRepos {
		fmt.Println(r.Repo)

		repoURL := repos.RepoURL{}
		if err := repoURL.Set(r.Repo); err != nil {
			panic(err)
		}
		if err := repoURL.ValidGitHubUrl(); err != nil {
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
		repoResult := pkg.RunScorecards(ctx, sugar, repoURL, checks.AllChecks)
		if err := repoResult.AsJSON( /*showDetails=*/ true, result); err != nil {
			panic(err)
		}
		//nolint
		logger.Sync() // flushes buffer, if any
	}
	result.Close()

	// copying the file to the GCS bucket
	if err := exec.Command("gsutil", "cp", fileName, fmt.Sprintf("gs://%s", os.Getenv("GCS_BUCKET"))).Run(); err != nil {
		panic(err)
	}
	//copying the results to the latest.json
	//nolint
	if err := exec.Command("gsutil", "cp", fmt.Sprintf("gs://%s/%s", os.Getenv("GCS_BUCKET"), fileName),
		fmt.Sprintf("gs://%s/latest.json", os.Getenv("GCS_BUCKET"))).Run(); err != nil {
		panic(err)
	}

	fmt.Println("Finished")
}
