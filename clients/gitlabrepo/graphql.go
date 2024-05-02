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

package gitlabrepo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

type graphqlHandler struct {
	err         error
	client      *http.Client
	graphClient *graphql.Client
	ctx         context.Context
	repourl     *repoURL
}

func (handler *graphqlHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.err = nil

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITLAB_AUTH_TOKEN")},
	)
	handler.client = oauth2.NewClient(ctx, src)
	handler.graphClient = graphql.NewClient(fmt.Sprintf("%s/api/graphql", repourl.Host()), handler.client)
}

type graphqlMRData struct {
	Project struct {
		MergeRequests struct {
			Nodes []graphqlMergeRequestNode `graphql:"nodes"`
		} `graphql:"mergeRequests(sort: MERGED_AT_DESC, state: merged, mergedBefore: $mergedBefore)"`
	} `graphql:"project(fullPath: $fullPath)"`
	QueryComplexity struct {
		Limit int `graphql:"limit"`
		Score int `graphql:"score"`
	} `graphql:"queryComplexity"`
}

type graphqlMergeRequestNode struct {
	MergedAt       time.Time `graphql:"mergedAt"`
	IID            string    `graphql:"iid"`
	MergeCommitSHA string    `graphql:"mergeCommitSha"`
	Author         struct {
		Username string    `graphql:"username"`
		ID       GitlabGID `graphql:"id"`
	} `graphql:"author"`
	MergedBy struct {
		Username string    `graphql:"username"`
		ID       GitlabGID `graphql:"id"`
	} `graphql:"mergeUser"`
	ID      GitlabGID `graphql:"id"`
	Commits struct {
		Nodes []struct {
			SHA string `graphql:"sha"`
		} `graphql:"nodes"`
	} `graphql:"commits"`
	Reviewers struct {
		Nodes []struct {
			Username                string `graphql:"username"`
			MergeRequestInteraction struct {
				ReviewState string `graphql:"reviewState"`
			} `graphql:"mergeRequestInteraction"`
			ID GitlabGID `graphql:"id"`
		} `graphql:"nodes"`
	} `graphql:"reviewers"`
	Approvers struct {
		Nodes []struct {
			Username string    `graphql:"username"`
			ID       GitlabGID `graphql:"id"`
		} `graphql:"nodes"`
	} `graphql:"approvedBy"`
}

type graphqlSBOMData struct {
	Project graphqlProject `graphql:"project(fullPath: $fullPath)"`
}

type graphqlProject struct {
	Pipelines graphqlPipelines `graphql:"pipelines(ref: $defaultBranch, first: 20)"`
}

type graphqlPipelines struct {
	Nodes []graphqlPipelineNode
}

type graphqlPipelineNode struct {
	Status       string               `graphql:"status"`
	JobArtifacts []graphqlJobArtifact `graphql:"jobArtifacts"`
}

type graphqlJobArtifact struct {
	Name         string `graphql:"name"`
	FileType     string `graphql:"fileType"`
	DownloadPath string `graphql:"downloadPath"`
}

type GitlabGID struct {
	Type string
	ID   int
}

var errGitlabID = errors.New("failed to parse gitlab id")

func (g *GitlabGID) UnmarshalJSON(data []byte) error {
	re := regexp.MustCompile(`gid:\/\/gitlab\/(\w+)\/(\d+)`)
	m := re.FindStringSubmatch(string(data))
	if len(m) < 3 {
		return fmt.Errorf("%w: %s", errGitlabID, string(data))
	}
	g.Type = m[1]

	id, err := strconv.Atoi(m[2])
	if err != nil {
		return fmt.Errorf("gid parse error: %w", err)
	}
	g.ID = id

	return nil
}

func (handler *graphqlHandler) getMergeRequestsDetail(before *time.Time) (graphqlMRData, error) {
	data := graphqlMRData{}
	path := fmt.Sprintf("%s/%s", handler.repourl.owner, handler.repourl.project)
	params := map[string]interface{}{
		"fullPath":     path,
		"mergedBefore": before,
	}
	err := handler.graphClient.Query(context.Background(), &data, params)
	if err != nil {
		return graphqlMRData{}, fmt.Errorf("couldn't query gitlab graphql for merge requests: %w", err)
	}

	return data, nil
}

func (handler *graphqlHandler) getSBOMDetail() (graphqlSBOMData, error) {
	data := graphqlSBOMData{}
	path := fmt.Sprintf("%s/%s", handler.repourl.owner, handler.repourl.project)
	params := map[string]interface{}{
		"fullPath":      path,
		"defaultBranch": graphql.String(handler.repourl.defaultBranch),
	}
	err := handler.graphClient.Query(context.Background(), &data, params)
	if err != nil {
		return graphqlSBOMData{}, fmt.Errorf("couldn't query gitlab graphql for SBOM Detail: %w", err)
	}

	return data, nil
}
