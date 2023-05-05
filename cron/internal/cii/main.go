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

// Package main implements the PubSub controller.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/cron/config"
	"github.com/ossf/scorecard/v4/cron/data"
)

const ciiBaseURL = "https://bestpractices.coreinfrastructure.org/projects.json"

type ciiPageResp struct {
	RepoURL    string `json:"repo_url"`
	BadgeLevel string `json:"badge_level"`
}

func writeToCIIDataBucket(ctx context.Context, pageResp []ciiPageResp, bucketURL string) error {
	for _, project := range pageResp {
		projectURL := strings.TrimPrefix(project.RepoURL, "https://")
		projectURL = strings.TrimPrefix(projectURL, "http://")
		jsonData, err := clients.BadgeResponse{
			BadgeLevel: project.BadgeLevel,
		}.AsJSON()
		if err != nil {
			return fmt.Errorf("error during AsJSON: %w", err)
		}
		fmt.Printf("Writing result for: %s\n", projectURL)
		if err := data.WriteToBlobStore(ctx, bucketURL,
			fmt.Sprintf("%s/result.json", projectURL), jsonData); err != nil {
			return fmt.Errorf("error during data.WriteToBlobStore: %w", err)
		}
	}
	return nil
}

func getPage(ctx context.Context, pageNum int) ([]ciiPageResp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s?page=%d", ciiBaseURL, pageNum), nil)
	if err != nil {
		return nil, fmt.Errorf("error during http.NewRequestWithContext: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error during http.DefaultClient.Do: %w", err)
	}
	defer resp.Body.Close()

	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error during io.ReadAll: %w", err)
	}

	ciiResponse := []ciiPageResp{}
	if err := json.Unmarshal(respContent, &ciiResponse); err != nil {
		return nil, fmt.Errorf("error during json.Unmarshal: %w", err)
	}
	return ciiResponse, nil
}

func main() {
	ctx := context.Background()
	fmt.Println("Starting...")

	flag.Parse()
	if err := config.ReadConfig(); err != nil {
		panic(err)
	}

	ciiDataBucket, err := config.GetCIIDataBucketURL()
	if err != nil {
		panic(err)
	}

	pageNum := 1
	pageResp, err := getPage(ctx, pageNum)
	for err == nil && len(pageResp) > 0 {
		if err := writeToCIIDataBucket(ctx, pageResp, ciiDataBucket); err != nil {
			panic(err)
		}
		pageNum++
		pageResp, err = getPage(ctx, pageNum)
	}
	if err != nil {
		panic(err)
	}

	fmt.Println("Job completed")
}
