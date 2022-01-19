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
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checks"
)

type scorecard struct {
	Repo   string `json:"Repo"`
	Date   string `json:"Date"`
	Checks []struct {
		CheckName  string   `json:"CheckName"`
		Details    []string `json:"Details"`
		Confidence int      `json:"Confidence"`
		Pass       bool     `json:"Pass"`
	} `json:"Checks"`
	MetaData []string `json:"MetaData"`
}

var _ = Describe("E2E TEST:executable", func() {
	Context("E2E TEST:Validating executable test", func() {
		It("Should return valid test results for scorecard", func() {
			file, err := os.ReadFile("../output/results.json")
			Expect(err).Should(BeNil())

			data := scorecard{}

			err = json.Unmarshal(file, &data)
			Expect(err).Should(BeNil())

			Expect(len(data.MetaData)).ShouldNot(BeZero())
			Expect(data.MetaData[0]).Should(BeEquivalentTo("openssf"))

			for _, c := range data.Checks {
				switch c.CheckName {
				case checks.CheckMaintained:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckDependencyUpdateTool:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckBranchProtection:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckCITests:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckCIIBestPractices:
					Expect(c.Pass).Should(BeFalse(), c.CheckName)
				case checks.CheckCodeReview:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckContributors:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckPinnedDependencies:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckFuzzing:
					Expect(c.Pass).Should(BeFalse(), c.CheckName)
				case checks.CheckPackaging:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckSAST:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckSecurityPolicy:
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case checks.CheckSignedReleases:
					Expect(c.Confidence).ShouldNot(Equal(10))
					Expect(c.Pass).Should(BeFalse(), c.CheckName)
				}
			}
		})
	})
})
