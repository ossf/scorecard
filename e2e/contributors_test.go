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

package e2e

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/gitlabrepo"
	scut "github.com/ossf/scorecard/v4/utests"
)

var _ = Describe("E2E TEST:"+checks.CheckContributors, func() {
	Context("E2E TEST:Validating project contributors", func() {
		It("Should return valid project contributors", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			}
			result := checks.Contributors(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "several contributors", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})

		It("Should return valid project contributors - GitLab", func() {
			repo, err := gitlabrepo.MakeGitlabRepo("https://gitlab.com/fdroid/fdroidclient")
			Expect(err).ShouldNot(BeNil())

			client, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), "", repo.Host())
			Expect(err).ShouldNot(BeNil())

			err = client.InitRepo(repo, "HEAD", 20)
			Expect(err).ShouldNot(BeNil())

			c, err := client.ListContributors()
			// Authentication is failing when querying users, not sure yet how to get around that
			if err != nil {
				errMsg := fmt.Sprintf("%v", err)

				if !(strings.Contains(errMsg, "error during Users.Get") && strings.Contains(errMsg, "401")) {
					Fail(fmt.Sprintf("couldn't list gitlab repo contributors: %v", err))
				}
			}
			Expect(len(c)).Should(BeNumerically(">", 0))
		})
	})
})
