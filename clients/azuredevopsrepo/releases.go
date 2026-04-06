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

package azuredevopsrepo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

type releasesHandler struct {
	gitClient git.Client
	ctx       context.Context
	errSetup  error
	once      *sync.Once
	repourl   *Repo
	getRefs   fnGetRefs
	releases  []clients.Release
}

func (r *releasesHandler) init(ctx context.Context, repourl *Repo) {
	r.ctx = ctx
	r.repourl = repourl
	r.errSetup = nil
	r.once = new(sync.Once)
	r.releases = nil
	r.getRefs = r.gitClient.GetRefs
}

func (r *releasesHandler) setup() error {
	r.once.Do(func() {
		filter := "tags/"
		peelTags := true
		response, err := r.getRefs(r.ctx, git.GetRefsArgs{
			RepositoryId: &r.repourl.id,
			Filter:       &filter,
			PeelTags:     &peelTags,
		})
		if err != nil {
			r.errSetup = fmt.Errorf("error getting tag refs: %w", err)
			return
		}

		r.releases = make([]clients.Release, 0, len(response.Value))
		for i := range response.Value {
			ref := &response.Value[i]
			if ref.Name == nil {
				continue
			}

			tagName := strings.TrimPrefix(*ref.Name, "refs/tags/")

			// For annotated tags, PeeledObjectId points to the commit.
			// For lightweight tags, ObjectId is the commit directly.
			commitish := ""
			if ref.PeeledObjectId != nil && *ref.PeeledObjectId != "" {
				commitish = *ref.PeeledObjectId
			} else if ref.ObjectId != nil {
				commitish = *ref.ObjectId
			}

			r.releases = append(r.releases, clients.Release{
				TagName:         tagName,
				TargetCommitish: commitish,
			})
		}
	})

	return r.errSetup
}

func (r *releasesHandler) listReleases() ([]clients.Release, error) {
	if err := r.setup(); err != nil {
		return nil, fmt.Errorf("error during releasesHandler.setup: %w", err)
	}

	return r.releases, nil
}
