// Copyright 2023 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E TEST: gitlabrepo.graphqlHandler", func() {
	var graphqlhandler graphqlHandler

	Context("E2E TEST: Confirm query result - GitLab", func() {
		It("Should have sufficient number of merge requests", func() {
			repo, err := MakeGitlabRepo("gitlab.com/gitlab-org/gitlab")
			Expect(err).Should(BeNil())

			graphqlhandler.init(context.Background(), repo.(*repoURL))
			data := graphqlData{}

			path := fmt.Sprintf("%s/%s", graphqlhandler.repourl.owner, graphqlhandler.repourl.project)
			params := map[string]interface{}{
				"fullPath":     path,
				"mergedBefore": time.Now(),
			}
			err = graphqlhandler.graphClient.Query(context.Background(), &data, params)
			Expect(err).Should(BeNil())

			Expect(len(data.Project.MergeRequests.Nodes)).Should(BeNumerically(">=", 100))
		})
	})

	Context("E2E TEST: Validate query cost - GitLab", func() {
		It("Should not have increased for HEAD query", func() {
			repo, err := MakeGitlabRepo("gitlab.com/gitlab-org/gitlab")
			Expect(err).Should(BeNil())

			graphqlhandler.init(context.Background(), repo.(*repoURL))
			data := graphqlData{}

			path := fmt.Sprintf("%s/%s", graphqlhandler.repourl.owner, graphqlhandler.repourl.project)
			params := map[string]interface{}{
				"fullPath":     path,
				"mergedBefore": time.Now(),
			}
			err = graphqlhandler.graphClient.Query(context.Background(), &data, params)
			Expect(err).Should(BeNil())

			Expect(data.QueryComplexity.Limit).Should(BeNumerically(">=", 200))
			Expect(data.QueryComplexity.Score).Should(BeNumerically("<=", 75))
		})
	})
})
