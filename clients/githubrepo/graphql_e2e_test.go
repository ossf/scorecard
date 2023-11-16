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
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/log"
)

var _ = Describe("E2E TEST: githubrepo.graphqlHandler", func() {
	var graphqlhandler *graphqlHandler

	BeforeEach(func() {
		graphqlhandler = &graphqlHandler{
			client: graphClient,
		}
	})

	Context("E2E TEST: Confirm Paging Commits Works", func() {
		It("Should only have 1 commit", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			vars := map[string]interface{}{
				"owner":                  githubv4.String("ossf"),
				"name":                   githubv4.String("scorecard"),
				"pullRequestsToAnalyze":  githubv4.Int(1),
				"issuesToAnalyze":        githubv4.Int(30),
				"issueCommentsToAnalyze": githubv4.Int(30),
				"reviewsToAnalyze":       githubv4.Int(30),
				"labelsToAnalyze":        githubv4.Int(30),
				"commitsToAnalyze":       githubv4.Int(1),
				"commitExpression":       githubv4.String("heads/main"),
				"historyCursor":          (*githubv4.String)(nil),
			}
			ctx := context.Background()
			logger := log.NewLogger(log.DebugLevel)
			rt := roundtripper.NewTransport(ctx, logger)
			httpClient := &http.Client{
				Transport: rt,
			}
			graphClient := githubv4.NewClient(httpClient)
			handler := &graphqlHandler{
				client: graphClient,
			}
			handler.init(context.Background(), repourl, 1)
			commits, err := populateCommits(handler, vars)
			Expect(err).To(BeNil())
			Expect(len(commits)).Should(BeEquivalentTo(1))
		})
		It("Should have 30 commits", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			vars := map[string]interface{}{
				"owner":                  githubv4.String("ossf"),
				"name":                   githubv4.String("scorecard"),
				"pullRequestsToAnalyze":  githubv4.Int(1),
				"issuesToAnalyze":        githubv4.Int(30),
				"issueCommentsToAnalyze": githubv4.Int(30),
				"reviewsToAnalyze":       githubv4.Int(30),
				"labelsToAnalyze":        githubv4.Int(30),
				"commitsToAnalyze":       githubv4.Int(30),
				"commitExpression":       githubv4.String("heads/main"),
				"historyCursor":          (*githubv4.String)(nil),
			}
			ctx := context.Background()
			logger := log.NewLogger(log.DebugLevel)
			rt := roundtripper.NewTransport(ctx, logger)
			httpClient := &http.Client{
				Transport: rt,
			}
			graphClient := githubv4.NewClient(httpClient)
			handler := &graphqlHandler{
				client: graphClient,
			}
			handler.init(context.Background(), repourl, 30)
			commits, err := populateCommits(handler, vars)
			Expect(err).To(BeNil())
			Expect(len(commits)).Should(BeEquivalentTo(30))
		})
		It("Should have 101 commits", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			vars := map[string]interface{}{
				"owner":                  githubv4.String("ossf"),
				"name":                   githubv4.String("scorecard"),
				"pullRequestsToAnalyze":  githubv4.Int(1),
				"issuesToAnalyze":        githubv4.Int(30),
				"issueCommentsToAnalyze": githubv4.Int(30),
				"reviewsToAnalyze":       githubv4.Int(30),
				"labelsToAnalyze":        githubv4.Int(30),
				"commitsToAnalyze":       githubv4.Int(101),
				"commitExpression":       githubv4.String("heads/main"),
				"historyCursor":          (*githubv4.String)(nil),
			}
			ctx := context.Background()
			logger := log.NewLogger(log.DebugLevel)
			rt := roundtripper.NewTransport(ctx, logger)
			httpClient := &http.Client{
				Transport: rt,
			}
			graphClient := githubv4.NewClient(httpClient)
			handler := &graphqlHandler{
				client: graphClient,
			}
			handler.init(context.Background(), repourl, 101)
			commits, err := populateCommits(handler, vars)
			Expect(err).To(BeNil())
			Expect(len(commits)).Should(BeEquivalentTo(101))
		})
	})

	Context("E2E TEST: Validate query cost", func() {
		It("Should not have increased for HEAD query", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			graphqlhandler.init(context.Background(), repourl, 30)
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
			graphqlhandler.init(context.Background(), repourl, 30)
			Expect(graphqlhandler.setup()).Should(BeNil())
			Expect(graphqlhandler.data).ShouldNot(BeNil())
			Expect(graphqlhandler.data.RateLimit.Cost).ShouldNot(BeNil())
			Expect(*graphqlhandler.data.RateLimit.Cost).Should(BeNumerically("<=", 1))
		})
	})
})
