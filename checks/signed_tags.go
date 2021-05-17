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
	"errors"

	"github.com/ossf/scorecard/checker"
	"github.com/shurcooL/githubv4"
)

const (
	signedTagsStr = "Signed-Tags"
	tagLookBack   = 5
)

func init() {
	registerCheck(signedTagsStr, SignedTags)
}

func SignedTags(c *checker.CheckRequest) checker.CheckResult {
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
		return checker.MakeRetryResult(signedTagsStr, err)
	}
	totalTags := 0
	totalSigned := 0
	for _, t := range query.Repository.Refs.Nodes {
		sha := string(t.Target.Oid)
		totalTags++
		gt, _, err := c.Client.Git.GetTag(c.Ctx, c.Owner, c.Repo, sha)
		if err != nil {
			c.Logf("!! unable to find the annotated commit: %s", sha)
			continue
		}
		if gt.GetVerification().GetVerified() {
			c.Logf("verified tag found: %s, commit: %s", t.Name, sha)
			totalSigned++
		} else {
			c.Logf("!! unverified tag found: %s, commit: %s, reason: %s", t.Name, sha, gt.GetVerification().GetReason())
		}
	}

	if totalTags == 0 {
		c.Logf("no tags found")
		return checker.MakeInconclusiveResult(signedTagsStr, errors.New("no signed tags found"))
	}

	c.Logf("found %d out of %d verified tags", totalSigned, totalTags)
	return checker.MakeProportionalResult(signedTagsStr, totalSigned, totalTags, 0.8)
}
