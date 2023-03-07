// Copyright 2022 OpenSSF Scorecard Authors
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
	"strings"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type releasesHandler struct {
	glClient *gitlab.Client
	once     *sync.Once
	errSetup error
	repourl  *repoURL
	releases []clients.Release
}

func (handler *releasesHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *releasesHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListReleases only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}
		releases, _, err := handler.glClient.Releases.ListReleases(handler.repourl.project, &gitlab.ListReleasesOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("%w: ListReleases failed", err)
			return
		}
		if len(releases) > 0 {
			handler.releases = releasesFrom(releases)
		} else {
			handler.releases = nil
		}
	})
	return handler.errSetup
}

func (handler *releasesHandler) getReleases() ([]clients.Release, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during Releases.setup: %w", err)
	}
	return handler.releases, nil
}

func releasesFrom(data []*gitlab.Release) []clients.Release {
	var releases []clients.Release
	for _, r := range data {
		release := clients.Release{
			TagName:         r.TagName,
			URL:             r.Assets.Links[0].DirectAssetURL,
			TargetCommitish: r.CommitPath,
		}
		for _, a := range r.Assets.Sources {
			release.Assets = append(release.Assets, clients.ReleaseAsset{
				Name: a.Format,
				URL:  a.URL,
			})
		}
		releases = append(releases, release)
	}
	return releases
}
