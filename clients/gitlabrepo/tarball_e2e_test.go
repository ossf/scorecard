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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E TEST: gitlabrepo.ListFiles", func() {
	Context("Test List Files- GitLab", func() {
		It("returns list of files", func() {
			repo, err := MakeGitlabRepo("https://gitlab.com/ossf-test/scorecard-pipeline-example")
			Expect(err).Should(BeNil())

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())

			err = client.InitRepo(repo, "HEAD", 0)
			Expect(err).Should(BeNil())

			files, err := client.ListFiles(func(s string) (bool, error) {
				return true, nil
			})
			Expect(err).Should(BeNil())
			Expect(len(files)).ShouldNot(BeZero())

			data, err := client.GetFileContent("README.md")
			Expect(err).Should(BeNil())
			Expect(len(data)).ShouldNot(BeZero())
		})
	})
})
