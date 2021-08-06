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
//nolint:dupl
package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v2/checker"
	"github.com/ossf/scorecard/v2/checks"
	"github.com/ossf/scorecard/v2/clients/githubrepo"
	scut "github.com/ossf/scorecard/v2/utests"
)

var _ = Describe("E2E TEST:"+checks.CheckTokenPermissions, func() {
	Context("E2E TEST:Validating token permission check", func() {
		It("Should return token permission works", func() {
			dl := scut.TestDetailLogger{}
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), ghClient, graphClient)
			err := repoClient.InitRepo("ossf-tests", "scorecard-check-token-permissions-e2e")
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:         context.Background(),
				Client:      ghClient,
				HTTPClient:  httpClient,
				RepoClient:  repoClient,
				Owner:       "ossf-tests",
				Repo:        "scorecard-check-token-permissions-e2e",
				GraphClient: graphClient,
				Dlogger:     &dl,
			}
			expected := scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			}
			result := checks.TokenPermissions(&req)
			// UPGRADEv2: to remove.
			// Old version.

			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeFalse())
			// New version.
			Expect(scut.ValidateTestReturn(nil, "token permissions", &expected, &result, &dl)).Should(BeTrue())
		})
	})
})
