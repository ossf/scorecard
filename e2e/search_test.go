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

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/ossfuzz"
)

var _ = Describe("E2E TEST:Search", func() {
	Context("E2E TEST:Search", func() {
		It("Should return valid hits", func() {
			fuzzClient := ossfuzz.CreateOSSFuzzClient(ossfuzz.StatusURL)
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			req := clients.SearchRequest{
				Query:    repoClient.URI(),
				Filename: "project.yaml",
			}
			resp, err := fuzzClient.Search(req)
			Expect(err).Should(BeNil())
			Expect(resp.Hits).Should(BeNumerically(">", 0))
		})
		It("Should return 0 hits", func() {
			fuzzClient := ossfuzz.CreateOSSFuzzClient(ossfuzz.StatusURL)
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			req := clients.SearchRequest{
				Query:    "notfound",
				Filename: "notfound.yaml",
			}
			resp, err := fuzzClient.Search(req)
			Expect(err).Should(BeNil())
			Expect(resp.Hits).Should(BeNumerically("==", 0))
		})
	})
})
