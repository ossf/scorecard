// Copyright 2025 OpenSSF Scorecard Authors
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

	"github.com/google/go-github/v82/github"

	sce "github.com/ossf/scorecard/v5/errors"
)

// IsReleaseImmutable reports whether the release tagged tag in owner/repo is
// published with GitHub's immutable-release guarantee. This lets workflows
// that reference an action via a tag (e.g. `owner/repo@vX.Y.Z`) be treated as
// pinned when the referenced release is immutable, per GitHub's guidance:
// https://docs.github.com/en/actions/how-tos/create-and-publish-actions/using-immutable-releases-and-tags-to-manage-your-actions-releases
//
// If the repository or tag doesn't exist, or the release is not immutable
// (e.g. immutable releases aren't enabled, the ref isn't a release tag, or
// the release predates the setting being enabled), this returns false with
// no error.
//
// When scorecard is configured against a GitHub Enterprise Server via
// GH_HOST, GitHub Actions resolves `uses:` references using an
// enterprise-first lookup and falls back to github.com through GitHub
// Connect. IsReleaseImmutable mirrors that: it queries the configured host
// first, and if the repository (or release) can't be resolved there, falls
// back to a separately authenticated github.com client (when one is
// configured).
func (client *Client) IsReleaseImmutable(owner, repo, tag string) (bool, error) {
	immutable, found, err := queryReleaseImmutable(client.ctx, client.repoClient, owner, repo, tag)
	if err != nil {
		return false, err
	}
	if found {
		return immutable, nil
	}

	if client.dotcomClient == nil {
		return false, nil
	}

	immutable, _, err = queryReleaseImmutable(client.ctx, client.dotcomClient, owner, repo, tag)
	if err != nil {
		return false, err
	}
	return immutable, nil
}

// queryReleaseImmutable fetches the release tagged tag in owner/repo from
// ghClient via the "get a release by tag name" REST API and reports whether
// it's immutable. found reports whether the repository/tag could be
// resolved on the queried host at all; a 404 is treated as "not found"
// rather than an error so callers can fall back to another host.
func queryReleaseImmutable(
	ctx context.Context,
	ghClient *github.Client,
	owner, repo, tag string,
) (immutable, found bool, err error) {
	release, resp, err := ghClient.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, false, nil
		}
		return false, false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetReleaseByTag: %v", err))
	}
	return release.GetImmutable(), true, nil
}
