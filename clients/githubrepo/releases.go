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

package githubrepo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type releasesHandler struct {
	client   *github.Client
	once     *sync.Once
	ctx      context.Context
	errSetup error
	repourl  *repoURL
	releases []clients.Release
}

func (handler *releasesHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.releases = nil
}

func (handler *releasesHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListReleases only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}
		releases, _, err := handler.client.Repositories.ListReleases(
			handler.ctx, handler.repourl.owner, handler.repourl.repo, &github.ListOptions{})
		if err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
		}
		handler.releases = releasesFrom(releases)
	})
	return handler.errSetup
}

func (handler *releasesHandler) getReleases() ([]clients.Release, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during graphqlHandler.setup: %w", err)
	}
	return handler.releases, nil
}

func releasesFrom(data []*github.RepositoryRelease) []clients.Release {
	var releases []clients.Release
	for _, r := range data {
		release := clients.Release{
			TagName:         r.GetTagName(),
			URL:             r.GetURL(),
			TargetCommitish: r.GetTargetCommitish(),
		}
		for _, a := range r.Assets {
			release.Assets = append(release.Assets, clients.ReleaseAsset{
				Name: a.GetName(),
				URL:  a.GetURL(),
			})
		}
		releases = append(releases, release)
	}
	return releases
}
