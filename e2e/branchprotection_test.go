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

	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
	sce "github.com/ossf/scorecard/errors"
	scut "github.com/ossf/scorecard/utests"
)

var _ = Describe("E2E TEST:Branch Protection", func() {
	Context("E2E TEST:Validating branch protection", func() {
		It("Should fail to return branch protection on other repositories", func() {
			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				Ctx:         context.Background(),
				Client:      ghClient,
				HTTPClient:  httpClient,
				RepoClient:  nil,
				Owner:       "apache",
				Repo:        "airflow",
				GraphClient: graphClient,
				Dlogger:     &dl,
			}
			expected := scut.TestReturn{
				Errors:        []error{sce.ErrScorecardInternal},
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.BranchProtection(&req)
			// UPGRADEv2: to remove.
			// Old version.
			Expect(result.Error).ShouldNot(BeNil())
			Expect(result.Pass).Should(BeFalse())
			// New version.
			Expect(scut.ValidateTestReturn(nil, "branch protection not accessible", &expected, &result, &dl)).Should(BeTrue())
		})
	})
})
