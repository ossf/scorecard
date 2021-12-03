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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/checks"
	"github.com/ossf/scorecard/v3/clients/githubrepo"
	scut "github.com/ossf/scorecard/v3/utests"
)

var _ = Describe("E2E TEST:"+checks.LicenseCheckPolicy, func() {
	Context("E2E TEST:Validating license file check", func() {
		It("Should return license check works", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-dangerous-workflow-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo)
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
			result := checks.LicenseCheck(&req)

			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())

			Expect(scut.ValidateTestReturn(nil, "license check", &expected, &result,
				&dl)).Should(BeTrue())
		})
	})
})
