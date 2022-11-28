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
	"net/http"
	"path"
	"sync"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
)

type licensesHandler struct {
	ghclient *github.Client
	once     *sync.Once
	ctx      context.Context
	errSetup error
	repourl  *repoURL
	licenses []clients.License
}

func (handler *licensesHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

// TODO: Can add support to parse the raw response JSON and mark licenses that are not in
// our defined License consts in clients/licenses.go as "not supported licenses".
func (handler *licensesHandler) setup() error {
	handler.once.Do(func() {
		client := handler.ghclient
		// defined at docs.github.com/en/rest/licenses#get-the-license-for-a-repository
		reqURL := path.Join("repos", handler.repourl.owner, handler.repourl.repo, "license")
		req, err := client.NewRequest("GET", reqURL, nil)
		if err != nil {
			handler.errSetup = fmt.Errorf("request for repo license failed with %w", err)
			return
		}
		bodyJSON := github.RepositoryLicense{}
		// The client.repoClient.Do API writes the response body to var bodyJSON,
		// so we can ignore the first returned variable (the entire http response object)
		// since we only need the response body here.
		resp, derr := client.Do(handler.ctx, req, &bodyJSON)
		switch resp.StatusCode {
		// Handle 400 error, perhaps the API changed.
		case http.StatusBadRequest:
			handler.errSetup = fmt.Errorf("bad request for repo license code %d, %w", resp.StatusCode, derr)
			return
		// Handle 404 error, appears that the repo has no license,
		// just return no need to log or error off.
		case http.StatusNotFound:
			return
		}
		if derr != nil {
			handler.errSetup = fmt.Errorf("response for repo license failed with %w", derr)
			return
		}

		// TODO: github.RepositoryLicense{} only supports one license per repo
		//       should that change to an array of licenses, the change would
		//       be here to iterate over any such range.
		handler.licenses = append(handler.licenses, clients.License{
			Key:    bodyJSON.GetLicense().GetKey(),
			Name:   bodyJSON.GetLicense().GetName(),
			SPDXId: bodyJSON.GetLicense().GetSPDXID(),
			Path:   bodyJSON.GetName(),
			Type:   bodyJSON.GetType(),
			Size:   bodyJSON.GetSize(),
		},
		)
		handler.errSetup = nil
	})

	return handler.errSetup
}

func (handler *licensesHandler) listLicenses() ([]clients.License, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during licensesHandler.setup: %w", err)
	}
	return handler.licenses, nil
}
