// Copyright 2025 OpenSSF Scorecard Authors
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
)

var _ = Describe("E2E TEST: githubrepo.tagsHandler", func() {
	var tagshandler *tagsHandler

	BeforeEach(func() {
		tagshandler = &tagsHandler{
			graphClient: graphClient,
		}
	})
	Context("E2E TEST: getTag", func() {
		It("Should return a tag", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")

			repourl := &Repo{
				owner: "AdamKorcz",
				repo:  "scorecard-testing",
			}
			tagshandler.init(context.Background(), repourl)
			err := tagshandler.setup()
			Expect(err).Should(BeNil())
			tagRef, err := tagshandler.getTag("test-tag1")
			Expect(err).Should(BeNil())
			Expect(tagRef).ShouldNot(BeNil())
			Expect(tagRef.ProtectionRule).ShouldNot(BeNil())
			Expect(tagRef.ProtectionRule.AllowForcePushes).ShouldNot(BeNil())
			Expect(tagRef.ProtectionRule.AllowDeletions).ShouldNot(BeNil())
			Expect(tagRef.ProtectionRule.RequireLinearHistory).ShouldNot(BeNil())
			Expect(tagRef.ProtectionRule.EnforceAdmins).ShouldNot(BeNil())
			Expect(tagRef.ProtectionRule.RequireLastPushApproval).Should(BeNil())
			Expect(tagRef.ProtectionRule.PullRequestRule.Required).To(Equal(asPtr(false)))
			Expect(tagRef.ProtectionRule.PullRequestRule.DismissStaleReviews).Should(BeNil())
			Expect(tagRef.ProtectionRule.PullRequestRule.RequireCodeOwnerReviews).Should(BeNil())
			Expect(tagRef.ProtectionRule.AllowDeletions).To(Equal(asPtr(false)))
			Expect(tagRef.ProtectionRule.AllowForcePushes).To(Equal(asPtr(false)))
		})
	})
})
