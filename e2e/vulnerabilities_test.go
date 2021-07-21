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

//nolint:dupl // repeating test cases that are slightly different is acceptable
package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
	scut "github.com/ossf/scorecard/utests"
)

var _ = Describe("E2E TEST:Vulnerabilities", func() {
	Context("E2E TEST:Validating vulnerabilities status", func() {
		It("Should return that there are no vulnerabilities", func() {
			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				Ctx:         context.Background(),
				Client:      ghClient,
				HTTPClient:  httpClient,
				RepoClient:  nil,
				Owner:       "ossf",
				Repo:        "scorecard",
				GraphClient: graphClient,
				Dlogger:     &dl,
			}
			expected := scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.HasUnfixedVulnerabilities(&req)
			// UPGRADEv2: to remove.
			// Old version.
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
			// New version.
			Expect(scut.ValidateTestReturn(nil, "no osv vulnerabilities", &expected, &result, &dl)).Should(BeTrue())
		})

		It("Should return that there are vulnerabilities", func() {
			dl := scut.TestDetailLogger{}
			checkRequest := checker.CheckRequest{
				Ctx:         context.Background(),
				Client:      ghClient,
				HTTPClient:  httpClient,
				RepoClient:  nil,
				Owner:       "oliverchang",
				Repo:        "open62541",
				GraphClient: graphClient,
				Dlogger:     &dl,
			}
			expected := scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.HasUnfixedVulnerabilities(&checkRequest)
			// UPGRADEv2: to remove.
			// Old version.
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeFalse())
			// New version.
			Expect(scut.ValidateTestReturn(nil, "osv vulnerabilities", &expected, &result, &dl)).Should(BeTrue())
		})
	})
})
