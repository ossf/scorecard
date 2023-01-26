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

package pkg

import (
	"context"
	"sort"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	sclog "github.com/ossf/scorecard/v4/log"
)

func (s *ScorecardResult) normalize() {
	sort.Slice(s.Checks, func(i, j int) bool {
		return s.Checks[i].Name < s.Checks[j].Name
	})
}

func countDetails(c []checker.CheckDetail) (debug, info, warn int) {
	for i := range c {
		switch c[i].Type {
		case checker.DetailDebug:
			debug++
		case checker.DetailInfo:
			info++
		case checker.DetailWarn:
			warn++
		}
	}
	return debug, info, warn
}

func compareScorecardResults(a, b *ScorecardResult) bool {
	if a.Repo != b.Repo || a.Scorecard != b.Scorecard || len(a.Checks) != len(b.Checks) {
		return false
	}

	for i := range a.Checks {
		if a.Checks[i].Name != b.Checks[i].Name ||
			a.Checks[i].Version != b.Checks[i].Version ||
			a.Checks[i].Score != b.Checks[i].Score ||
			a.Checks[i].Reason != b.Checks[i].Reason {
			return false
		}

		// details are only compared using the number of debug, info and warn
		aDebug, aInfo, aWarn := countDetails(a.Checks[i].Details)
		bDebug, bInfo, bWarn := countDetails(b.Checks[i].Details)
		if aDebug != bDebug || aInfo != bInfo || aWarn != bWarn {
			return false
		}
	}
	return true
}

var _ = Describe("E2E TEST: RunScorecard with re-used repoClient", func() {
	Context("E2E TEST: Validate results are identical regardless of order", func() {
		assertLastResultsIdentical := func(repos []string) {
			if len(repos) < 2 {
				return
			}
			ctx := context.Background()
			isolatedLogger := sclog.NewLogger(sclog.DefaultLevel)
			lastRepo := repos[len(repos)-1]
			repo, rc, ofrc, cc, vc, err := checker.GetClients(ctx, lastRepo, "", isolatedLogger)
			Expect(err).Should(BeNil())
			isolatedResult, err := RunScorecard(ctx, repo, clients.HeadSHA, 0, checks.GetAll(), rc, ofrc, cc, vc)
			Expect(err).Should(BeNil())
			logger := sclog.NewLogger(sclog.DefaultLevel)
			_, rc, ofrc, cc, vc, err = checker.GetClients(ctx, repos[0], "", logger)
			Expect(err).Should(BeNil())

			var sharedResult ScorecardResult
			for i := range repos {
				repo, err = githubrepo.MakeGithubRepo(repos[i])
				Expect(err).Should(BeNil())
				sharedResult, err = RunScorecard(ctx, repo, clients.HeadSHA, 0, checks.GetAll(), rc, ofrc, cc, vc)
				Expect(err).Should(BeNil())
			}

			isolatedResult.normalize()
			sharedResult.normalize()
			Expect(isolatedResult).To(BeComparableTo(sharedResult, cmp.Comparer(compareScorecardResults)))
		}
		It("A then B results should be produce the same distribution of details as the isolated B results", func() {
			assertLastResultsIdentical([]string{
				"https://github.com/ossf-tests/scorecard",
				"https://github.com/ossf-tests/scorecard-action",
			})
		})
	})
})
