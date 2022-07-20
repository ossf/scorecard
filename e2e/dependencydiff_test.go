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
	"github.com/ossf/scorecard/v4/pkg"

	sclog "github.com/ossf/scorecard/v4/log"
)

const (
	OWNER = "ossf-tests"
	REPO  = "scorecard-depdiff"
	BASE  = "fd2a82b3b735fffbc2d782ed5f50301b879ecc51"
	HEAD  = "1989568f93e484f6a86f8b276b170e3d6962ce12"
)

var _ = Describe("E2E TEST:"+dependencydiff.Depdiff, func() {
	Context("E2E TEST:Validating use of the dependency-diff API", func() {
		It("Should return a slice of dependency-diff checking results", func() {
			ctx := context.Background()
			logger := sclog.NewLogger(sclog.DefaultLevel)
			ownerName, repoName := OWNER, REPO
			baseSHA, headSHA := BASE, HEAD
			scorecardChecksNames := []string{
				checks.CheckBranchProtection,
			}
			changeTypesToCheck := map[pkg.ChangeType]bool{
				pkg.Removed: true, // Only checking those removed ones will make this test faster.
			}
			results, err := dependencydiff.GetDependencyDiffResults(
				ctx, logger,
				ownerName, repoName, baseSHA, headSHA,
				scorecardChecksNames,
				changeTypesToCheck,
			)
			Expect(err).Should(BeNil())
			Expect(len(results) > 0).Should(BeTrue())
		})
		It("Should return a valid empty result", func() {
			ctx := context.Background()
			ownerName, repoName := OWNER, REPO
			baseSHA, headSHA := BASE, BASE

			scorecardChecksNames := []string{
				checks.CheckBranchProtection,
			}
			changeTypesToCheck := map[pkg.ChangeType]bool{
				pkg.Removed: true,
			}
			results, err := dependencydiff.GetDependencyDiffResults(
				ctx, logger,
				ownerName, repoName, baseSHA, headSHA,
				scorecardChecksNames,
				changeTypesToCheck,
			)
			Expect(err).Should(BeNil())
			Expect(len(results) == 0).Should(BeTrue())
		})
		It("Should initialize clients corresponding to the checks to run and do not crash", func() {
			ctx := context.Background()
			ownerName, repoName := OWNER, REPO
			baseSHA, headSHA := BASE, HEAD

			scorecardChecksNames := []string{}
			changeTypesToCheck := map[pkg.ChangeType]bool{
				pkg.Removed: true,
			}
			_, err := dependencydiff.GetDependencyDiffResults(
				ctx, logger,
				ownerName, repoName, baseSHA, headSHA,
				scorecardChecksNames,
				changeTypesToCheck,
			)
			Expect(err).Should(BeNil())
		})
	})
})
