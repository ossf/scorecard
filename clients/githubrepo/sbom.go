// Copyright 2024 OpenSSF Scorecard Authors
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

	"github.com/google/go-github/v53/github"

	"github.com/ossf/scorecard/v5/clients"
)

type SBOMHandler struct {
	ghclient *github.Client
	once     *sync.Once
	ctx      context.Context
	errSetup error
	repourl  *repoURL
	SBOMs    []clients.SBOM
}

func (handler *SBOMHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.SBOMs = nil
}

func (handler *SBOMHandler) setup() error {
	handler.once.Do(func() {
		// Check for SBOMs in pipeline artifacts
		err := handler.checkCICDArtifacts()
		if err != nil {
			handler.errSetup = fmt.Errorf("failed searching for SBOM in CICD artifacts: %w", err)
		}
	})

	return handler.errSetup
}

func (handler *SBOMHandler) listSBOMs() ([]clients.SBOM, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during SBOMHandler.setup: %w", err)
	}

	return handler.SBOMs, nil
}

func (handler *SBOMHandler) checkCICDArtifacts() error {
	// Originally wanted to use workflowruns from latest release, but
	// that would've resulted in as many api calls as workflows runs (11 for scorcard itself)
	// Seems like deficiency in github api (or my understanding of it)
	client := handler.ghclient
	// defined at: (using apiVersion=2022-11-28)
	// docs.github.com/en/rest/actions/artifacts?#list-artifacts-for-a-repository
	reqURL := path.Join("repos", handler.repourl.owner, handler.repourl.repo, "actions", "artifacts")
	req, err := client.NewRequest("GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("request for repo artifacts failed with %w", err)
	}
	bodyJSON := github.ArtifactList{}
	// The client.repoClient.Do API writes the response body to var bodyJSON,
	// so we can ignore the first returned variable (the entire http response object)
	// since we only need the response body here.
	resp, derr := client.Do(handler.ctx, req, &bodyJSON)
	if derr != nil {
		return fmt.Errorf("response for repo SBOM failed with %w", derr)
	}

	if resp.StatusCode != http.StatusOK {
		// Dont fail, just return
		// TODO: print info for users that a non-200 response was returned
		return nil
	}

	if bodyJSON.GetTotalCount() == 0 {
		return nil
	}

	returnedArtifacts := bodyJSON.Artifacts

	for i := range returnedArtifacts {
		artifact := returnedArtifacts[i]

		if *artifact.Expired || !clients.ReSBOMFile.MatchString(artifact.GetName()) {
			continue
		}

		handler.SBOMs = append(handler.SBOMs, clients.SBOM{
			Name:    artifact.GetName(),
			URL:     artifact.GetArchiveDownloadURL(),
			Created: artifact.CreatedAt.Time,
			Path:    artifact.GetURL(),
		},
		)
	}

	return nil
}
