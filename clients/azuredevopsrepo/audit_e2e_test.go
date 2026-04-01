// Copyright 2026 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v5/clients"
)

var _ = Describe("E2E TEST: azuredevopsrepo.auditHandler", func() {
	Context("Audit - Azure DevOps", func() {
		It("Should return repository creation date", func() {
			repo, err := MakeAzureDevOpsRepo("https://dev.azure.com/openssf-scorecard/scorecard-testing/_git/scorecard-testing")
			Expect(err).Should(BeNil())

			repoClient, err := CreateAzureDevOpsClient(context.Background(), repo)
			Expect(err).Should(BeNil())

			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())

			createdAt, err := repoClient.GetCreatedAt()
			// GetCreatedAt may fail if the PAT lacks audit log permissions,
			// but falls back to first commit date. Either way, it should not error.
			Expect(err).Should(BeNil())
			Expect(createdAt).ShouldNot(Equal(time.Time{}))

			Expect(repoClient.Close()).Should(BeNil())
		})
	})
})
