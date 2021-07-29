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

	"github.com/ossf/scorecard/v2/checker"
	"github.com/ossf/scorecard/v2/checks"
	scut "github.com/ossf/scorecard/v2/utests"
)

var _ = Describe("E2E TEST:SAST", func() {
	Context("E2E TEST:Validating use of SAST tools", func() {
		It("Should return use of SAST tools", func() {
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
				Errors:        nil,
				Score:         7,
				NumberOfWarn:  1,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			}
			result := checks.SAST(&req)
			// UPGRADEv2: to remove.
			// Old version.
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeFalse())
			// New version.
			Expect(scut.ValidateTestReturn(nil, "sast used", &expected, &result, &dl)).Should(BeTrue())
		})
	})
})
