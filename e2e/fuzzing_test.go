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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/checks/raw"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/ossfuzz"
	scut "github.com/ossf/scorecard/v4/utests"
)

var _ = Describe("E2E TEST:"+checks.CheckFuzzing, func() {
	Context("E2E TEST:Validating use of fuzzing tools", func() {
		It("Should return use of OSS-Fuzz", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("tensorflow/tensorflow")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			ossFuzzRepoClient, err := ossfuzz.CreateOSSFuzzClientEager(ossfuzz.StatusURL)
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:         context.Background(),
				RepoClient:  repoClient,
				OssFuzzRepo: ossFuzzRepoClient,
				Repo:        repo,
				Dlogger:     &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Fuzzing(&req)
			Expect(scut.ValidateTestReturn(nil, "use fuzzing", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
			Expect(ossFuzzRepoClient.Close()).Should(BeNil())
		})
		It("Should return use of ClusterFuzzLite", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-fuzzing-cflite")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			ossFuzzRepoClient, err := ossfuzz.CreateOSSFuzzClientEager(ossfuzz.StatusURL)
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:         context.Background(),
				RepoClient:  repoClient,
				OssFuzzRepo: ossFuzzRepoClient,
				Repo:        repo,
				Dlogger:     &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Fuzzing(&req)
			Expect(scut.ValidateTestReturn(nil, "use fuzzing", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
			Expect(ossFuzzRepoClient.Close()).Should(BeNil())
		})
		It("Should return use of GoBuiltInFuzzers", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-fuzzing-golang")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			ossFuzzRepoClient, err := ossfuzz.CreateOSSFuzzClientEager(ossfuzz.StatusURL)
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:         context.Background(),
				RepoClient:  repoClient,
				OssFuzzRepo: ossFuzzRepoClient,
				Repo:        repo,
				Dlogger:     &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			}
			result := checks.Fuzzing(&req)
			Expect(scut.ValidateTestReturn(nil, "use fuzzing", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
			Expect(ossFuzzRepoClient.Close()).Should(BeNil())
		})
		It("Should return an expected number of GoBuiltInFuzzers", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-fuzzing-golang")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			ossFuzzRepoClient, err := ossfuzz.CreateOSSFuzzClientEager(ossfuzz.StatusURL)
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:         context.Background(),
				RepoClient:  repoClient,
				OssFuzzRepo: ossFuzzRepoClient,
				Repo:        repo,
				Dlogger:     &dl,
			}
			rawData, err := raw.Fuzzing(&req)
			Expect(err).Should(BeNil())
			Expect(len(rawData.Fuzzers) == 1).Should(BeTrue())
		})
		It("Should return no fuzzing", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-packaging-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			ossFuzzRepoClient, err := ossfuzz.CreateOSSFuzzClientEager(ossfuzz.StatusURL)
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:         context.Background(),
				RepoClient:  repoClient,
				OssFuzzRepo: ossFuzzRepoClient,
				Repo:        repo,
				Dlogger:     &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Fuzzing(&req)
			Expect(scut.ValidateTestReturn(nil, "no fuzzing", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
			Expect(ossFuzzRepoClient.Close()).Should(BeNil())
		})
	})
})
