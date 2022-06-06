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

package githubrepo

import (
	"context"
	"fmt"
	"path"

	"github.com/google/go-github/v38/github"
)

type languagesHandler struct {
	ghclient *github.Client
	ctx      context.Context
	repourl  *repoURL
}

func (handler *languagesHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
}

func (handler *languagesHandler) listProgrammingLanguages() (map[string]int, error) {
	client := handler.ghclient
	reqURL := path.Join("repos", handler.repourl.owner, handler.repourl.repo, "languages")
	req, err := client.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request for repo languages failed with %w", err)
	}
	bodyJSON := map[string]int{}
	// The client.repoClient.Do API writes the reponse body to var bodyJSON,
	// so we can ignore the first returned variable (the http response object)
	// since we only need the response body here.
	_, errResp := client.Do(handler.ctx, req, &bodyJSON)
	if errResp != nil {
		return nil, fmt.Errorf("response for repo languages failed with %w", err)
	}
	return bodyJSON, nil
}
