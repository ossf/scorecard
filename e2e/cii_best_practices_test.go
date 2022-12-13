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

var _ = Describe("E2E TEST:"+checks.CheckCIIBestPractices, func() {
	Context("E2E TEST:Validating use of CII Best Practices", func() {
		It("Should return use of CII Best Practices", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("tensorflow/tensorflow")
			Expect(err).Should(BeNil())
			ciiClient := clients.DefaultCIIBestPracticesClient()

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: nil,
				CIIClient:  ciiClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         5,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.CIIBestPractices(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "passing badge", &expected, &result, &dl)).Should(BeTrue())
		})
	})
})
