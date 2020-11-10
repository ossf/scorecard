// Copyright 2020 Security Scorecard Authors
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

package checks

import (
	"github.com/ossf/scorecard/checker"
	"github.com/shurcooL/githubv4"
)

var tagLookBack int = 5

func init() {
	registerCheck("Signed-Tags", SignedTags)
}

func SignedTags(c checker.Checker) checker.CheckResult {

	type ref struct {
		Name   githubv4.String
		Target struct {
			Oid githubv4.String
		}
	}
	var query struct {
		Repository struct {
			Refs struct {
				Nodes []ref
			} `graphql:"refs(refPrefix: \"refs/tags/\", last: $count)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(c.Owner),
		"name":  githubv4.String(c.Repo),
		"count": githubv4.Int(tagLookBack),
	}

	if err := c.GraphClient.Query(c.Ctx, &query, variables); err != nil {
		return checker.RetryResult(err)
	}

	totalReleases := 0
	totalSigned := 0
	for _, t := range query.Repository.Refs.Nodes {
		sha := string(t.Target.Oid)
		totalReleases++
		gt, _, err := c.Client.Git.GetTag(c.Ctx, c.Owner, c.Repo, sha)
		if err != nil {
			return checker.RetryResult(err)
		}
		if gt.GetVerification().GetVerified() {
			c.Logf("signed tag found: %s, commit: %s", t.Name, sha)
			totalSigned++
		}
	}

	return checker.ProportionalResult(totalSigned, totalReleases, 0.8)
}
