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

	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

type releasesHandler struct {
	client   *github.Client
	once     *sync.Once
	ctx      context.Context
	errSetup error
	repourl  *Repo
	releases []clients.Release
}

func (handler *releasesHandler) init(ctx context.Context, repourl *Repo) {
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
		tags, err := handler.listTagReleases()
		if err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
			return
		}
		handler.releases = mergeTagReleases(handler.releases, tags)
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
				URL:  r.GetHTMLURL(),
			})
		}
		releases = append(releases, release)
	}
	return releases
}

func (handler *releasesHandler) listTagReleases() ([]clients.Release, error) {
	refs, _, err := handler.client.Git.ListMatchingRefs(
		handler.ctx, handler.repourl.owner, handler.repourl.repo, "tags/",
	)
	if err != nil {
		return nil, fmt.Errorf("Git.ListMatchingRefs: %w", err)
	}

	releases := make([]clients.Release, 0, len(refs))
	for _, ref := range refs {
		tag := handler.releaseTagFromRef(ref)
		if tag == nil {
			continue
		}

		releases = append(releases, clients.Release{
			TagName:         tag.Name,
			URL:             tag.URL,
			TargetCommitish: tag.TargetCommitish,
			Tag:             tag,
		})
	}
	return releases, nil
}

func (handler *releasesHandler) releaseTagFromRef(ref *github.Reference) *clients.ReleaseTag {
	if ref == nil || ref.Object == nil {
		return nil
	}

	name := strings.TrimPrefix(ref.GetRef(), "refs/tags/")
	if name == "" {
		return nil
	}

	tag := &clients.ReleaseTag{
		Name:            name,
		URL:             ref.GetURL(),
		TargetCommitish: ref.Object.GetSHA(),
	}

	if ref.Object.GetType() != "tag" {
		return tag
	}

	gitTag, _, err := handler.client.Git.GetTag(
		handler.ctx, handler.repourl.owner, handler.repourl.repo, ref.Object.GetSHA(),
	)
	if err != nil {
		return tag
	}

	if gitTag.Object != nil {
		tag.TargetCommitish = gitTag.Object.GetSHA()
	}
	if gitTag.Verification != nil {
		tag.SignatureVerified = gitTag.Verification.GetVerified()
	}
	return tag
}

func mergeTagReleases(releases, tags []clients.Release) []clients.Release {
	if len(tags) == 0 {
		return releases
	}

	releaseByTag := make(map[string]int, len(releases))
	for i := range releases {
		releaseByTag[releases[i].TagName] = i
	}

	for i := range tags {
		tagRelease := tags[i]
		if releaseIndex, ok := releaseByTag[tagRelease.TagName]; ok {
			releases[releaseIndex].Tag = tagRelease.Tag
			if releases[releaseIndex].TargetCommitish == "" {
				releases[releaseIndex].TargetCommitish = tagRelease.TargetCommitish
			}
			continue
		}
		releases = append(releases, tagRelease)
	}
	return releases
}
