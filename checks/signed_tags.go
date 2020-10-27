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
	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

var tagLookBack int = 5

func init() {
	registerCheck("Signed-Tags", SignedTags)
}

func SignedTags(c checker.Checker) checker.CheckResult {
	tags, _, err := c.Client.Repositories.ListTags(c.Ctx, c.Owner, c.Repo, &github.ListOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	totalReleases := 0
	totalSigned := 0
	for _, t := range tags {
		totalReleases++
		gt, _, err := c.Client.Git.GetCommit(c.Ctx, c.Owner, c.Repo, t.GetCommit().GetSHA())
		if err != nil {
			return checker.RetryResult(err)
		}
		if gt.GetVerification().GetVerified() {
			c.Logf("signed tag found: %s, commit: %s", *t.Name, t.GetCommit().GetSHA())
			totalSigned++
		}
		if totalReleases > tagLookBack {
			break
		}
	}

	return checker.ProportionalResult(totalSigned, totalReleases, 0.8)
}
