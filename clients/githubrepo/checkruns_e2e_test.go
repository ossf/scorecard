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

	"github.com/google/go-github/v53/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/log"
)

var _ = Describe("E2E TEST: githubrepo.checkrunsHandler", func() {
	var checkrunshandler *checkrunsHandler

	BeforeEach(func() {
		checkrunshandler = &checkrunsHandler{
			graphClient: graphClient,
		}
	})

	// TODO: Add e2e tests for commit depth.

	Context("E2E TEST: Validate query cost", func() {
		It("Should not have increased query cost", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			checkrunshandler.init(context.Background(), repourl, 30)
			Expect(checkrunshandler.setup()).Should(BeNil())
			Expect(checkrunshandler.checkData).ShouldNot(BeNil())
			Expect(checkrunshandler.checkData.RateLimit.Cost).ShouldNot(BeNil())
			Expect(*checkrunshandler.checkData.RateLimit.Cost).Should(BeNumerically("<=", 1))
		})
	})
	Context("E2E TEST: listCheckRunsForRef", func() {
		It("Should return check runs for a valid ref", func() {
			repourl := &repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}
			rt := roundtripper.NewTransport(context.Background(), &log.Logger{})
			httpClient := &http.Client{
				Transport: rt,
			}
			checkrunshandler.init(context.Background(), repourl, 30)
			checkrunshandler.client = github.NewClient(httpClient)
			Expect(checkrunshandler.setup()).Should(BeNil())
			Expect(checkrunshandler.checkData).ShouldNot(BeNil())

			checkRuns, err := checkrunshandler.listCheckRunsForRef("main")
			Expect(err).Should(BeNil())
			Expect(checkRuns).ShouldNot(BeNil())
			Expect(len(checkRuns)).Should(BeNumerically(">", 0))
		})
	})
	It("Should return an error for an invalid ref", func() {
		repourl := &repoURL{
			owner:     "ossf",
			repo:      "scorecard",
			commitSHA: clients.HeadSHA,
		}
		checkrunshandler.init(context.Background(), repourl, 30)
		Expect(checkrunshandler.setup()).Should(BeNil())
		rt := roundtripper.NewTransport(context.Background(), &log.Logger{})
		httpClient := &http.Client{
			Transport: rt,
		}
		checkrunshandler.client = github.NewClient(httpClient)
		checkRuns, err := checkrunshandler.listCheckRunsForRef("invalid-ref")
		Expect(err).ShouldNot(BeNil())
		Expect(checkRuns).Should(BeNil())
	})
})
