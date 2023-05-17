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

package githubrepo

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/clients"
)

var _ = Describe("E2E TEST: githubrepo.branchesHandler", func() {
	var brancheshandler *branchesHandler

	BeforeEach(func() {
		brancheshandler = &branchesHandler{
			graphClient: graphClient,
		}
	})

	Context("E2E TEST: Validate query cost", func() {
		It("Should not have increased for HEAD query", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")

			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			brancheshandler.init(context.Background(), repourl)
			Expect(brancheshandler.setup()).Should(BeNil())
			Expect(brancheshandler.data).ShouldNot(BeNil())
			Expect(brancheshandler.data.RateLimit.Cost).ShouldNot(BeNil())
			Expect(*brancheshandler.data.RateLimit.Cost).Should(BeNumerically("<=", 1))
		})

		It("Should fail for non-HEAD query", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")

			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: "de5224bbc56eceb7a25aece55d2d53bbc561ed2d",
			}
			brancheshandler.init(context.Background(), repourl)
			Expect(brancheshandler.setup()).ShouldNot(BeNil())
			Expect(brancheshandler.data).Should(BeNil())
		})
	})

	Context("E2E TEST: Get default branch", func() {
		It("Should return the correct default branch", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")

			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			brancheshandler.init(context.Background(), repourl)

			branchRef, err := brancheshandler.getDefaultBranch()
			Expect(err).Should(BeNil())
			Expect(branchRef).ShouldNot(BeNil())
			Expect(*branchRef.Name).Should(Equal("main"))
		})
	})
	Context("E2E TEST: getBranch", func() {
		It("Should return a branch", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")

			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			brancheshandler.init(context.Background(), repourl)

			branchRef, err := brancheshandler.getBranch("main")
			Expect(err).Should(BeNil())
			Expect(branchRef).ShouldNot(BeNil())
		})

		It("Should return an error for non-existent branch", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")

			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			brancheshandler.init(context.Background(), repourl)

			branchRef, err := brancheshandler.getBranch("non-existent-branch")
			Expect(err).Should(BeNil())
			Expect(branchRef).Should(BeNil())
		})
	})
	Context("E2E TEST: query branch", func() {
		It("Should return a branch", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")

			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			brancheshandler.init(context.Background(), repourl)
			Expect(brancheshandler.query("main")).ShouldNot(BeNil())
		})

		It("Should fail for non-HEAD query", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")

			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: "de5224bbc56eceb7a25aece55d2d53bbc561ed2d",
			}
			brancheshandler.init(context.Background(), repourl)
			branchRef, err := brancheshandler.query("main")

			Expect(err).ShouldNot(BeNil())
			Expect(branchRef).Should(BeNil())
		})
	})
})
