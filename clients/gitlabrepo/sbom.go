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

	"github.com/ossf/scorecard/v5/clients"
)

var (
	errSBOMDataEmpty        = errors.New("SBOM Data not found")
	errLatestPipelinesEmpty = errors.New("missing Sbom Pipelines Info")
)

type SBOMHandler struct {
	glClient *gitlab.Client
	once     *sync.Once
	errSetup error
	repourl  *repoURL
	sboms    []clients.SBOM
}

func (handler *SBOMHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *SBOMHandler) setup(sbomData graphqlSBOMData) error {
	handler.once.Do(func() {
		if cmp.Equal(sbomData, graphqlSBOMData{}) {
			handler.errSetup = errSBOMDataEmpty
			return
		}

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

func (handler *SBOMHandler) listSBOMs(sbomData graphqlSBOMData) ([]clients.SBOM, error) {
	if err := handler.setup(sbomData); err != nil {
		return nil, fmt.Errorf("error during SBOMHandler.setup: %w", err)
	}

	return handler.sboms, nil
}

func (handler *SBOMHandler) checkCICDArtifacts(latestPipelines []graphqlPipelineNode) {
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

			handler.sboms = append(handler.sboms, clients.SBOM{
				Name: artifact.Name,
				URL:  artifact.DownloadPath,
			})
		}
	}
}
