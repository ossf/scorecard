// Copyright 2021 Security Scorecard Authors
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
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/gitlabrepo"
	scut "github.com/ossf/scorecard/v4/utests"
)

var _ = Describe("E2E TEST:"+checks.CheckVulnerabilities, func() {
	Context("E2E TEST:Validating vulnerabilities status", func() {
		It("Should return that there are no vulnerabilities", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA)
			Expect(err).Should(BeNil())

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				Ctx:                   context.Background(),
				RepoClient:            repoClient,
				VulnerabilitiesClient: clients.DefaultVulnerabilitiesClient(),
				Repo:                  repo,
				Dlogger:               &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}

			result := checks.Vulnerabilities(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "no osv vulnerabilities", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})

		It("Should return that there are vulnerabilities", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-vulnerabilities-open62541")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA)
			Expect(err).Should(BeNil())

			dl := scut.TestDetailLogger{}
			checkRequest := checker.CheckRequest{
				Ctx:                   context.Background(),
				RepoClient:            repoClient,
				VulnerabilitiesClient: clients.DefaultVulnerabilitiesClient(),
				Repo:                  repo,
				Dlogger:               &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 3, // 3 vulnerabilities remove 3 points.
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Vulnerabilities(&checkRequest)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "osv vulnerabilities", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return that there are vulnerabilities at commit", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-vulnerabilities-open62541")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "de6367caa31b59e2156f83b04c2f30611b7ac393")
			Expect(err).Should(BeNil())

			dl := scut.TestDetailLogger{}
			checkRequest := checker.CheckRequest{
				Ctx:                   context.Background(),
				RepoClient:            repoClient,
				VulnerabilitiesClient: clients.DefaultVulnerabilitiesClient(),
				Repo:                  repo,
				Dlogger:               &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 3, // 3 vulnerabilities remove 3 points.
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Vulnerabilities(&checkRequest)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "osv vulnerabilities", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return that there are vulnerabilities - GitLab", func() {
			// project url is gitlab.com/N8BWert/scorecard-check-vulnerabilites-open62541.
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/N8BWert/39539557")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), os.Getenv("GITLAB_AUTH_TOKEN"), repo)
			Expect(err).Should(BeNil())
			err = repoClient.InitRepo(repo, clients.HeadSHA)
			Expect(err).Should(BeNil())

			dl := scut.TestDetailLogger{}
			checkRequest := checker.CheckRequest{
				Ctx:                   context.Background(),
				RepoClient:            repoClient,
				VulnerabilitiesClient: clients.DefaultVulnerabilitiesClient(),
				Repo:                  repo,
				Dlogger:               &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 3, // 3 vulnerabilities remove 3 points.
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Vulnerabilities(&checkRequest)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "osv vulnerabilities", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return that there are vulnerabilities at commit - GitLab", func() {
			// project url is gitlab.com/N8BWert/scorecard-check-vulnerabilites-open62541.
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/N8BWert/39539557")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), os.Getenv("GITLAB_AUTH_TOKEN"), repo)
			Expect(err).Should(BeNil())
			err = repoClient.InitRepo(repo, "de6367caa31b59e2156f83b04c2f30611b7ac393")
			Expect(err).Should(BeNil())

			dl := scut.TestDetailLogger{}
			checkRequest := checker.CheckRequest{
				Ctx:                   context.Background(),
				RepoClient:            repoClient,
				VulnerabilitiesClient: clients.DefaultVulnerabilitiesClient(),
				Repo:                  repo,
				Dlogger:               &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 3, // 3 vulnerabilities remove 3 points.
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Vulnerabilities(&checkRequest)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "osv vulnerabilities", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
	})
})
