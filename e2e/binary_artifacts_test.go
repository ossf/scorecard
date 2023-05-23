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

package e2e

import (
	"context"
	"os"

	"github.com/go-git/go-git/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/localdir"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TODO: use dedicated repo that don't change.
// TODO: need negative results.
var _ = Describe("E2E TEST:"+checks.CheckBinaryArtifacts, func() {
	Context("E2E TEST:Binary artifacts are not present in source code", func() {
		It("Should return not binary artifacts in source code", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			b := checks.BinaryArtifactsCheck{}
			result := b.Run(&req)
			Expect(scut.ValidateTestReturn(nil, "no binary artifacts", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return binary artifacts present in source code", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-binary-artifacts-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			// TODO: upload real binaries to the repo as well.
			// There are 24 dummy binaries that are ignoreed because they only contain ASCII characters.
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 1,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			b := checks.BinaryArtifactsCheck{}
			result := b.Run(&req)
			Expect(scut.ValidateTestReturn(nil, "binary artifacts", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return binary artifacts present at commit in source code", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-binary-artifacts-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "5b48dea88825662d67ed94b609b45cf7705333b6", 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			// TODO: upload real binaries to the repo as well.
			// There are 24 dummy binaries that are ignoreed because they only contain ASCII characters.
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 1,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			b := checks.BinaryArtifactsCheck{}
			result := b.Run(&req)
			Expect(scut.ValidateTestReturn(nil, "binary artifacts", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return no binary artifacts present in source code", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-binary-artifacts-e2e-4-binaries")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			// TODO: upload real binaries to the repo as well.
			// Existing binaries only contain SCII characters and are ignored.
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			b := checks.BinaryArtifactsCheck{}
			result := b.Run(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "binary artifacts", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return binary artifacts present at commit in source code", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-binary-artifacts-e2e-4-binaries")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "d994b3e1a8912283f9958a7c1e0aa480ca24a7ce", 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			// TODO: upload real binaries to the repo.
			// Existing binaries only contain SCII characters and are ignored.
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			b := checks.BinaryArtifactsCheck{}
			result := b.Run(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "binary artifacts", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return binary artifacts present at commit in source code when using local repoClient", func() {
			// create temp dir
			tmpDir, err := os.MkdirTemp("", "")
			Expect(err).Should(BeNil())
			defer os.RemoveAll(tmpDir)

			_, e := git.PlainClone(tmpDir, false, &git.CloneOptions{
				URL: "http://github.com/ossf-tests/scorecard-check-binary-artifacts-e2e-4-binaries",
			})
			Expect(e).Should(BeNil())

			dl := scut.TestDetailLogger{}
			repo, err := localdir.MakeLocalDirRepo(tmpDir)
			Expect(err).Should(BeNil())

			x := localdir.CreateLocalDirClient(context.Background(), logger)
			err = x.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: x,
				Repo:       repo,
				Dlogger:    &dl,
			}
			// TODO: upload real binaries to the repo.
			// Existing binaries only contain SCII characters and are ignored.
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			b := checks.BinaryArtifactsCheck{}
			result := b.Run(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "binary artifacts", &expected, &result, &dl)).Should(BeTrue())
			Expect(x.Close()).Should(BeNil())
		})
	})
})
