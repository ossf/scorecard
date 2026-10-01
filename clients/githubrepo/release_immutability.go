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
	"fmt"

	"github.com/shurcooL/githubv4"

	sce "github.com/ossf/scorecard/v5/errors"
)

//nolint:govet
type releaseImmutabilityData struct {
	Repository struct {
		Release struct {
			Immutable githubv4.Boolean
		} `graphql:"release(tagName: $tag)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

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
func (client *Client) IsReleaseImmutable(owner, repo, tag string) (bool, error) {
	var data releaseImmutabilityData
	vars := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(repo),
		"tag":   githubv4.String(tag),
	}
	if err := client.graphClient.client.Query(client.ctx, &data, vars); err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
	}
	return bool(data.Repository.Release.Immutable), nil
}
