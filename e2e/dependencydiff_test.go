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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/dependencydiff"
)

const (
	repoURI = "ossf-tests/scorecard-depdiff"
	base    = "fd2a82b3b735fffbc2d782ed5f50301b879ecc51"
	head    = "1989568f93e484f6a86f8b276b170e3d6962ce12"
)

// TODO (#2087): More e2e tests and a potnetial refactoring needed for the func getScorecardCheckResults.
var _ = Describe("E2E TEST:"+dependencydiff.Depdiff, func() {
	Context("E2E TEST:Validating use of the dependency-diff API", func() {
		It("Should return a slice of dependency-diff checking results", func() {
			ctx := context.Background()
			checksToRun := []string{
				checks.CheckBranchProtection,
			}
			changeTypesToCheck := []string{
				"removed", // Only checking those removed ones will make this test faster.
			}
			results, err := dependencydiff.GetDependencyDiffResults(
				ctx,
				repoURI,
				base, head,
				checksToRun,
				changeTypesToCheck,
			)
			Expect(err).Should(BeNil())
			Expect(len(results) > 0).Should(BeTrue())
		})
		It("Should return a valid empty result", func() {
			ctx := context.Background()
			checksToRun := []string{
				checks.CheckBranchProtection,
			}
			changeTypesToCheck := []string{
				"removed",
			}
			results, err := dependencydiff.GetDependencyDiffResults(
				ctx,
				repoURI,
				base, base,
				checksToRun,
				changeTypesToCheck,
			)
			Expect(err).Should(BeNil())
			Expect(len(results) == 0).Should(BeTrue())
		})
		It("Should initialize clients corresponding to the checks to run and do not crash", func() {
			ctx := context.Background()
			checksToRun := []string{
				checks.CheckFuzzing,
			}
			changeTypesToCheck := []string{
				"removed",
			}
			_, err := dependencydiff.GetDependencyDiffResults(
				ctx,
				repoURI,
				base, head,
				checksToRun,
				changeTypesToCheck,
			)
			Expect(err).Should(BeNil())
		})
	})
})
