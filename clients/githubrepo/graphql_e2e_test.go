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

package githubrepo

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/clients"
)

var _ = Describe("E2E TEST: githubrepo.graphqlHandler", func() {
	var graphqlhandler *graphqlHandler

	BeforeEach(func() {
		graphqlhandler = &graphqlHandler{
			client: graphClient,
		}
	})

	Context("E2E TEST: Validate query cost", func() {
		It("Should not have increased for HEAD query", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			graphqlhandler.init(context.Background(), repourl)
			Expect(graphqlhandler.setup()).Should(BeNil())
			Expect(graphqlhandler.data).ShouldNot(BeNil())
			Expect(graphqlhandler.data.RateLimit.Cost).ShouldNot(BeNil())
			Expect(*graphqlhandler.data.RateLimit.Cost).Should(BeNumerically("<=", 1))
		})
		It("Should not have increased for commit query", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: "de5224bbc56eceb7a25aece55d2d53bbc561ed2d",
			}
			graphqlhandler.init(context.Background(), repourl)
			Expect(graphqlhandler.setup()).Should(BeNil())
			Expect(graphqlhandler.data).ShouldNot(BeNil())
			Expect(graphqlhandler.data.RateLimit.Cost).ShouldNot(BeNil())
			Expect(*graphqlhandler.data.RateLimit.Cost).Should(BeNumerically("<=", 1))
		})
		It("Should not have increased for check run query", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			graphqlhandler.init(context.Background(), repourl)
			Expect(graphqlhandler.setupCheckRuns()).Should(BeNil())
			Expect(graphqlhandler.checkData).ShouldNot(BeNil())
			Expect(graphqlhandler.checkData.RateLimit.Cost).ShouldNot(BeNil())
			Expect(*graphqlhandler.checkData.RateLimit.Cost).Should(BeNumerically("<=", 1))
		})
	})
})
