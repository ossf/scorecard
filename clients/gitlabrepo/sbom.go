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

package gitlabrepo

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

var (
	errSbomDataEmpty        = errors.New("tarball not found")
	errReleaseNodesEmpty    = errors.New("corrupted tarball")
	errLatestPipelinesEmpty = errors.New("ZipSlip path detected")
)

type sbomHandler struct {
	glClient *gitlab.Client
	once     *sync.Once
	errSetup error
	repourl  *repoURL
	sboms    []clients.Sbom
}

func (handler *sbomHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *sbomHandler) setup(sbomData graphqlSbomData) error {
	handler.once.Do(func() {
		if cmp.Equal(sbomData, graphqlSbomData{}) {
			handler.errSetup = errSbomDataEmpty
			return
		}

		if len(sbomData.Project.Releases.Nodes) < 1 || sbomData.Project.Releases.Nodes[0].Assets.Links.Nodes == nil {
			handler.errSetup = errReleaseNodesEmpty
			return
		}

		// Check for sboms in release artifacts
		ReleaseAssetLinks := sbomData.Project.Releases.Nodes[0].Assets.Links.Nodes
		handler.checkReleaseArtifacts(ReleaseAssetLinks)

		latestPipelines := sbomData.Project.Pipelines.Nodes

		if latestPipelines == nil {
			handler.errSetup = errLatestPipelinesEmpty
			return
		}

		// Check for sboms in pipeline artifacts
		handler.checkCICDArtifacts(latestPipelines)

		handler.errSetup = nil
	})

	return handler.errSetup
}

func (handler *sbomHandler) listSboms(sbomData graphqlSbomData) ([]clients.Sbom, error) {
	if err := handler.setup(sbomData); err != nil {
		return nil, fmt.Errorf("error during sbomHandler.setup: %w", err)
	}

	return handler.sboms, nil
}

func (handler *sbomHandler) checkReleaseArtifacts(assetlinks []graphqlReleaseAssetLinksNode) {
	if len(assetlinks) < 1 { // no release links
		return
	}

	for _, link := range assetlinks {
		if !clients.ReSbomFile.Match([]byte(link.Name)) {
			continue
		}

		handler.sboms = append(handler.sboms, clients.Sbom{
			Name:   link.Name,
			URL:    link.URL,
			Path:   link.DirectAssetPath,
			Origin: "repositoryRelease",
		})
	}
}

func (handler *sbomHandler) checkCICDArtifacts(latestPipelines []graphqlPipelineNode) {
	// Checks latest 20 pipelines in default branch for appropriate artifacts
	for _, pipeline := range latestPipelines {
		if pipeline.Status != "SUCCESS" {
			continue
		}

		for _, artifact := range pipeline.JobArtifacts {
			if artifact.FileType != "CYCLONEDX" &&
				artifact.FileType != "DEPENDENCY_SCANNING" &&
				artifact.FileType != "CONTAINER_SCANNING" {
				continue
			}

			handler.sboms = append(handler.sboms, clients.Sbom{
				Name:   artifact.Name,
				URL:    artifact.DownloadPath,
				Origin: "repositoryCICD",
			})
		}
	}
}
