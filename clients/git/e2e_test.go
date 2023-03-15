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

package git

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/clients"
)

var _ = DescribeTable("Test ListCommits commit-depth for HEAD",
	func(uri string) {
		const commitSHA = clients.HeadSHA
		const commitDepth = 1
		client := &Client{}
		Expect(client.InitRepo(uri, commitSHA, commitDepth)).To(BeNil())
		commits, err := client.ListCommits()
		Expect(err).To(BeNil())
		Expect(len(commits)).Should(BeEquivalentTo(commitDepth))
		Expect(client.Close()).To(BeNil())
	},
	Entry("GitHub", "https://github.com/ossf/scorecard"),
	Entry("Local", "file://../../"),
	Entry("GitLab", "https://gitlab.haskell.org/haskell/filepath"),
)

var _ = DescribeTable("Test ListCommits commit-depth and latest commit at [0]",
	func(uri, commitSHA string) {
		const commitDepth = 10
		client := &Client{}
		Expect(client.InitRepo(uri, commitSHA, commitDepth)).To(BeNil())
		commits, err := client.ListCommits()
		Expect(err).To(BeNil())
		Expect(len(commits)).Should(BeEquivalentTo(commitDepth))
		Expect(commits[0].SHA).Should(BeEquivalentTo(commitSHA))
		Expect(client.Close()).To(BeNil())
	},
	Entry("GitHub", "https://github.com/ossf/scorecard", "c06ac740cc49fea404c54c036000731d5ea6ebe3"),
	Entry("Local", "file://../../", "c06ac740cc49fea404c54c036000731d5ea6ebe3"),
	Entry("GitLab", "https://gitlab.haskell.org/haskell/filepath", "98f8bba9eac8c7183143d290d319be7df76c258b"),
)

var _ = DescribeTable("Test ListCommits without enough commits",
	func(uri string) {
		const commitSHA = "dc1835b7ffe526969d65436b621e171e3386771e"
		const commitDepth = 10
		client := &Client{}
		Expect(client.InitRepo(uri, commitSHA, commitDepth)).To(BeNil())
		commits, err := client.ListCommits()
		Expect(err).To(BeNil())
		Expect(len(commits)).Should(BeEquivalentTo(3))
		Expect(commits[0].SHA).Should(BeEquivalentTo(commitSHA))
		Expect(client.Close()).To(BeNil())
	},
	Entry("GitHub", "https://github.com/ossf/scorecard"),
	Entry("Local", "file://../../"),
	// TODO(#1709): Add equivalent test for GitLab.
)

var _ = DescribeTable("Test Search across a repo",
	func(uri string) {
		const (
			commitSHA   = "c06ac740cc49fea404c54c036000731d5ea6ebe3"
			commitDepth = 10
		)
		client := &Client{}
		Expect(client.InitRepo(uri, commitSHA, commitDepth)).To(BeNil())
		resp, err := client.Search(clients.SearchRequest{
			Query: "github/codeql-action/analyze",
		})
		Expect(err).To(BeNil())
		Expect(resp.Hits).Should(BeNumerically(">=", 1))
		Expect(client.Close()).To(BeNil())
	},
	Entry("GitHub", "https://github.com/ossf/scorecard"),
	Entry("Local", "file://../../"),
	// TODO(#1709): Add equivalent test for GitLab.
)

var _ = DescribeTable("Test Search within a path",
	func(uri string) {
		const (
			commitSHA   = "c06ac740cc49fea404c54c036000731d5ea6ebe3"
			commitDepth = 10
		)
		client := &Client{}
		Expect(client.InitRepo(uri, commitSHA, commitDepth)).To(BeNil())
		resp, err := client.Search(clients.SearchRequest{
			Query: "github/codeql-action/analyze",
			Path:  ".github/workflows",
		})
		Expect(err).To(BeNil())
		Expect(resp.Hits).Should(BeEquivalentTo(1))
		Expect(client.Close()).To(BeNil())
	},
	Entry("GitHub", "https://github.com/ossf/scorecard"),
	Entry("Local", "file://../../"),
	// TODO(#1709): Add equivalent test for GitLab.
)

var _ = DescribeTable("Test Search within a filename",
	func(uri string) {
		const (
			commitSHA   = "c06ac740cc49fea404c54c036000731d5ea6ebe3"
			commitDepth = 10
		)
		client := &Client{}
		Expect(client.InitRepo(uri, commitSHA, commitDepth)).To(BeNil())
		resp, err := client.Search(clients.SearchRequest{
			Query:    "github/codeql-action/analyze",
			Filename: "codeql-analysis.yml",
		})
		Expect(err).To(BeNil())
		Expect(resp.Hits).Should(BeEquivalentTo(1))
		Expect(client.Close()).To(BeNil())
	},
	Entry("GitHub", "https://github.com/ossf/scorecard"),
	Entry("Local", "file://../../"),
	// TODO(#1709): Add equivalent test for GitLab.
)

var _ = DescribeTable("Test Search within path and filename",
	func(uri string) {
		const (
			commitSHA   = "c06ac740cc49fea404c54c036000731d5ea6ebe3"
			commitDepth = 10
		)
		client := &Client{}
		Expect(client.InitRepo(uri, commitSHA, commitDepth)).To(BeNil())
		resp, err := client.Search(clients.SearchRequest{
			Query:    "github/codeql-action/analyze",
			Path:     ".github/workflows",
			Filename: "codeql-analysis.yml",
		})
		Expect(err).To(BeNil())
		Expect(resp.Hits).Should(BeEquivalentTo(1))
		Expect(client.Close()).To(BeNil())
	},
	Entry("GitHub", "https://github.com/ossf/scorecard"),
	Entry("Local", "file://../../"),
	// TODO(#1709): Add equivalent test for GitLab.
)
