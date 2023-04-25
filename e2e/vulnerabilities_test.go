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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	scut "github.com/ossf/scorecard/v4/utests"
)

var _ = Describe("E2E TEST:"+checks.CheckVulnerabilities, func() {
	Context("E2E TEST:Validating vulnerabilities status", func() {
		It("Should return that there are vulnerabilities", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-vulnerabilities-open62541")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
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
				NumberOfWarn:  3,
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
			err = repoClient.InitRepo(repo, "de6367caa31b59e2156f83b04c2f30611b7ac393", 0)
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
				NumberOfWarn:  3,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Vulnerabilities(&checkRequest)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "osv vulnerabilities", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return that there are vulnerable packages", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-osv-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "2a81bfbc691786d6b8226a4092cca4f1509c842d", 0)
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
				Score:         checker.MaxResultScore - 2, // 2 vulnerabilities remove 2 points.
				NumberOfWarn:  2,
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
