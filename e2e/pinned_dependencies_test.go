// Copyright 2021 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package e2e

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/gitlabrepo"
	"github.com/ossf/scorecard/v4/clients/localdir"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TODO: use dedicated repo that don't change.
// TODO: need negative results.
var _ = Describe("E2E TEST:"+checks.CheckPinnedDependencies, func() {
	Context("E2E TEST:Validating dependencies check is working", func() {
		It("Should return dependencies check is working", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-pinned-dependencies-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  139,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			}
			result := checks.PinningDependencies(&req)
			Expect(scut.ValidateTestReturn(nil, "dependencies check", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return dependencies check at commit", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-pinned-dependencies-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "c8bfd7cf04ea7af741e1d07af98fabfcc1b6ffb1")
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  139,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			}
			result := checks.PinningDependencies(&req)
			Expect(scut.ValidateTestReturn(nil, "dependencies check", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return dependencies check for a local repoClient", func() {
			dl := scut.TestDetailLogger{}

			tmpDir, err := ioutil.TempDir("", "")
			Expect(err).Should(BeNil())
			defer os.RemoveAll(tmpDir)

			_, e := git.PlainClone(tmpDir, false, &git.CloneOptions{
				URL: "http://github.com/ossf-tests/scorecard-check-pinned-dependencies-e2e",
			})
			Expect(e).Should(BeNil())

			repo, err := localdir.MakeLocalDirRepo(tmpDir)
			Expect(err).Should(BeNil())

			x := localdir.CreateLocalDirClient(context.Background(), logger)
			err = x.InitRepo(repo, clients.HeadSHA)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: x,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  139,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			}
			result := checks.PinningDependencies(&req)
			Expect(scut.ValidateTestReturn(nil, "dependencies check", &expected, &result, &dl)).Should(BeTrue())
			Expect(x.Close()).Should(BeNil())
		})
		It("Should return dependencies check is working - GitLab", func() {
			dl := scut.TestDetailLogger{}
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.ossf.com/ossf-tests/scorecard-check-pinned-dependencies-e2e")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), os.Getenv("GITLAB_AUTH_TOKEN"), repo)
			Expect(err).Should(BeNil())
			err = repoClient.InitRepo(repo, clients.HeadSHA)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  139,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			}
			result := checks.PinningDependencies(&req)
			Expect(scut.ValidateTestReturn(nil, "dependencies check", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return dependencies check at commit - GitLab", func() {
			dl := scut.TestDetailLogger{}
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.ossf.com/ossf-tests/scorecard-check-pinned-dependencies-e2e")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), os.Getenv("GITLAB_AUTH_TOKEN"), repo)
			Expect(err).Should(BeNil())
			err = repoClient.InitRepo(repo, "c8bfd7cf04ea7af741e1d07af98fabfcc1b6ffb1")
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  139,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			}
			result := checks.PinningDependencies(&req)
			Expect(scut.ValidateTestReturn(nil, "dependencies check", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
	})
})
