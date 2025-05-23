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

package githubrepo

import (
	"context"
	"io"
	"net/http"

	"github.com/google/go-github/v53/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v5/internal/gitfile"
	"github.com/ossf/scorecard/v5/log"
)

var _ = Describe("E2E TEST: githubrepo.contributorsHandler", func() {
	var contribHandler *contributorsHandler
	var fileHandler io.ReadCloser
	repoURL := Repo{
		owner:     "ossf",
		repo:      "scorecard",
		commitSHA: clients.HeadSHA,
	}

	BeforeEach(func() {
		ctx := context.Background()
		rt := roundtripper.NewTransport(context.Background(), &log.Logger{})
		httpClient := &http.Client{
			Transport: rt,
		}
		client := github.NewClient(httpClient)
		contribHandler = &contributorsHandler{
			ghClient: client,
			ctx:      ctx,
		}
		//nolint:errcheck
		repo, _, _ := contribHandler.ghClient.Repositories.Get(contribHandler.ctx, repoURL.owner, repoURL.repo)
		gitFileHandler := gitfile.Handler{}
		gitFileHandler.Init(contribHandler.ctx, repo.GetCloneURL(), repoURL.commitSHA)
		// searching for CODEOWNERS file
		for _, path := range CodeOwnerPaths {
			var err error
			fileHandler, err = gitFileHandler.GetFile(path)
			if err == nil {
				break
			}
		}
	})
	Context("getContributors()", func() {
		skipIfTokenIsNot(patTokenType, "PAT only")
		It("returns contributors for valid HEAD query", func() {
			repoURL := Repo{
				owner:     "ossf-tests",
				repo:      "scorecard-check-contributors-e2e",
				commitSHA: clients.HeadSHA,
			}

			contribHandler.init(context.Background(), &repoURL)
			Expect(contribHandler.getContributors(fileHandler)).ShouldNot(BeNil())
			Expect(contribHandler.errSetup).Should(BeNil())
		})
	})
})
