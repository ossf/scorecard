// Copyright 2023 OpenSSF Scorecard Authors
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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E TEST: gitlabrepo.Branches", func() {
	Context("Test Default branch- GitLab", func() {
		It("returns branches for the repo", func() {
			skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")
			repo, err := MakeGitlabRepo("https://gitlab.com/ossf-test/scorecard")
			Expect(err).Should(BeNil())

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())

			err = client.InitRepo(repo, "HEAD", 0)
			Expect(err).Should(BeNil())
			branch, err := client.GetDefaultBranch()
			Expect(err).Should(BeNil())
			Expect(branch.Name).Should(Equal("main"))
		})
	})
})
