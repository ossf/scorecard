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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E TEST: gitlabrepo.ListProgrammingLanguages", func() {
	Context("Test list programming languages - GitLab", func() {
		It("returns branches for the repo", func() {
			repo, err := MakeGitlabRepo("https://gitlab.com/baserow/baserow")
			Expect(err).Should(BeNil())

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())

			err = client.InitRepo(repo, "HEAD", 0)
			Expect(err).Should(BeNil())
			programmingLang, err := client.ListProgrammingLanguages()
			Expect(err).Should(BeNil())
			Expect(programmingLang).ShouldNot(BeNil())
			// Check for the presence of some languages
			isPythonPresent := false
			for _, lang := range programmingLang {
				// compare case insensitive
				if strings.EqualFold(string(lang.Name), "Python") {
					isPythonPresent = true
					break
				}
			}
			Expect(isPythonPresent).Should(BeTrue())
		})
		It("Should return no programming languages repo with no code", func() {
			repo, err := MakeGitlabRepo("https://gitlab.com/ossf-test/scorecard-test-branches")
			Expect(err).Should(BeNil())

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())

			err = client.InitRepo(repo, "HEAD", 0)
			Expect(err).Should(BeNil())
			programmingLang, err := client.ListProgrammingLanguages()
			Expect(err).Should(BeNil())
			Expect(programmingLang).Should(BeNil())
		})
	})
})
