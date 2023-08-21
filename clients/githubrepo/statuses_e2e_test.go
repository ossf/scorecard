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
	"net/http"

	"github.com/google/go-github/v53/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/log"
)

var _ = Describe("E2E TEST: githubrepo.statusesHandler", func() {
	var statusHandler *statusesHandler

	BeforeEach(func() {
		ctx := context.Background()
		rt := roundtripper.NewTransport(context.Background(), &log.Logger{})
		httpClient := &http.Client{
			Transport: rt,
		}
		client := github.NewClient(httpClient)
		statusHandler = &statusesHandler{
			client: client,
			ctx:    ctx,
		}
	})
	Context("listStatuses()", func() {
		It("returns statuses", func() {
			repoURL := repoURL{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: clients.HeadSHA,
			}

			statusHandler.init(context.Background(), &repoURL)
			resp, err := statusHandler.listStatuses("8c9e552f68e5fd070692b8376ac51d2e8a7f0aaa")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp)).ShouldNot(Equal(0))
		})
	})
})
