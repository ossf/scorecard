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
	"fmt"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
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
		latestPipeline := sbomData.Project.Pipelines.Nodes
		ReleaseAssetLinks := sbomData.Project.Releases.Nodes[0].Assets.Links.Nodes

		// Check for sboms in release artifacts
		err := handler.checkReleaseArtifacts(ReleaseAssetLinks)
		if err != nil {
			handler.errSetup = fmt.Errorf("failed searching for Sbom in Release artifacts: %w", err)
			return
		}

		// Check for sboms in pipeline artifacts
		err = handler.checkCICDArtifacts(latestPipeline)
		if err != nil {
			handler.errSetup = fmt.Errorf("failed searching for Sbom in CICD artifacts: %w", err)
			return
		}

		handler.errSetup = nil
	})

	return handler.errSetup
}

func (handler *sbomHandler) listSboms(latestRelease graphqlSbomData) ([]clients.Sbom, error) {
	if err := handler.setup(latestRelease); err != nil {
		return nil, fmt.Errorf("error during sbomHandler.setup: %w", err)
	}

	return handler.sboms, nil
}

func (handler *sbomHandler) checkReleaseArtifacts(assetlinks []graphqlReleaseAssetLinksNode) error {
	if len(assetlinks) < 1 { // no release links
		return nil
	}

	for _, link := range assetlinks {
		if !clients.ReSbomFile.Match([]byte(link.Name)) {
			continue
		}

		handler.sboms = append(handler.sboms, clients.Sbom{
			Name:   link.Name,
			URL:    link.URL,
			Path:   link.DirectAssetPath,
			Origin: "repositoryAPI",
		})
	}

	return nil
}

func (handler *sbomHandler) checkCICDArtifacts(latestpipelines []graphqlPipelineNode) error {
	// Originally intended to check artifacts from latest release pipeline,
	// but changed to check latest default branch pipeline to align with github check
	for _, pipeline := range latestpipelines {
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
	return nil
}
