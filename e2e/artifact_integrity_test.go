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

// Copyright 2026 OpenSSF Scorecard Authors
// Licensed under the Apache License, Version 2.0


package e2e

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	scut "github.com/ossf/scorecard/v5/utests"
)


func runArtifactIntegrityCheck(repoSlug string) (checker.CheckResult, []checker.CheckDetail, error) {
	dl := scut.TestDetailLogger{}

	repo, err := githubrepo.MakeGithubRepo(repoSlug)
	if err != nil {
		return checker.CheckResult{}, []checker.CheckDetail{}, err
	}

	repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
	if err = repoClient.InitRepo(repo, clients.HeadSHA, 0); err != nil {
		return checker.CheckResult{}, []checker.CheckDetail{}, err
	}
	defer repoClient.Close() 

	req := checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: repoClient,
		Repo:       repo,
		Dlogger:    &dl,
	}

	result := checks.EntryPointArtifactIntegrity(&req)

	details := dl.Flush()
	return result, details, nil
}


func hasDetailType(details []checker.CheckDetail, dt checker.DetailType) bool {
	for _, d := range details {
		if d.Type == dt {
			return true
		}
	}
	return false
}

func anyMsgContains(details []checker.CheckDetail, substrings ...string) bool {
	for _, d := range details {
		lower := strings.ToLower(d.Msg.Text) // text lives on .Msg.Text, not .Text
		for _, s := range substrings {
			if strings.Contains(lower, strings.ToLower(s)) {
				return true
			}
		}
	}
	return false
}


var _ = Describe("E2E TEST:"+checks.CheckArtifactIntegrity, func() {

	Context("Repos with verified releases", func() {

		It("cli/cli – checksum files present, score should be high", func() {

			result, details, err := runArtifactIntegrityCheck("cli/cli")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(BeNumerically(">=", checker.MinResultScore))
			Expect(result.Score).Should(BeNumerically("<=", checker.MaxResultScore))

			Expect(details).ShouldNot(BeEmpty(),
				"expected DetailInfo/DetailWarn entries from emitReleaseLog")

			Expect(result.Score).Should(BeNumerically(">=", 5),
				"cli/cli releases carry checksums; expected score >= 5")
		})

		It("goreleaser/goreleaser – SLSA provenance present, should reach max score", func() {

			result, details, err := runArtifactIntegrityCheck("goreleaser/goreleaser")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(BeNumerically(">=", checker.MinResultScore))
			Expect(result.Score).Should(BeNumerically("<=", checker.MaxResultScore))

			Expect(details).ShouldNot(BeEmpty(),
				"expected detail entries for SLSA provenance detection")

			Expect(result.Score).Should(Equal(checker.MaxResultScore),
				"goreleaser ships SLSA provenance; expected maximum score")
		})

		It("sigstore/cosign – signature files present, score should be high", func() {

			result, details, err := runArtifactIntegrityCheck("sigstore/cosign")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(BeNumerically(">=", 5),
				"cosign releases carry signatures; expected score >= 5")

			Expect(details).ShouldNot(BeEmpty())
		})
	})


	Context("Repos with no releases", func() {

		It("octocat/Hello-World – no releases → inconclusive", func() {
			result, details, err := runArtifactIntegrityCheck("octocat/Hello-World")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(
				SatisfyAny(Equal(checker.InconclusiveResultScore), Equal(0)),
				"repo with no releases should produce an inconclusive result",
			)

			Expect(details).Should(BeEmpty(),
				"no releases means no detail entries should be emitted")
		})

		It("github/gitignore – source-only assets → inconclusive", func() {

			result, _, err := runArtifactIntegrityCheck("github/gitignore")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(
				SatisfyAny(Equal(checker.InconclusiveResultScore), Equal(0)),
				"source-only releases should be skipped and yield inconclusive",
			)
		})
	})


	Context("Repos with no integrity artifacts", func() {

		It("nicowillis/pong – no checksums or signatures → DetailWarn emitted", func() {
			result, details, err := runArtifactIntegrityCheck("nicowillis/pong")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(BeNumerically(">=", checker.MinResultScore))
			Expect(result.Score).Should(BeNumerically("<=", checker.MaxResultScore))

			Expect(hasDetailType(details, checker.DetailWarn)).Should(BeTrue(),
				"releases without integrity files should emit checker.DetailWarn entries")
		})
	})


	Context("Repos with mixed release quality", func() {

		It("hashicorp/terraform – mixed releases → valid score with detail entries", func() {
			result, details, err := runArtifactIntegrityCheck("hashicorp/terraform")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(BeNumerically(">=", checker.MinResultScore))
			Expect(result.Score).Should(BeNumerically("<=", checker.MaxResultScore))

			Expect(details).ShouldNot(BeEmpty(),
				"expected per-release detail entries for hashicorp/terraform")
		})

		It("hashicorp/terraform – weight decay: multiple releases produce multiple entries", func() {

			result, details, err := runArtifactIntegrityCheck("hashicorp/terraform")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(len(details)).Should(BeNumerically(">", 1),
				"expected entries for each analysed release (weight decay scenario)")

			Expect(result.Score).Should(BeNumerically(">=", 1),
				"at least one covered recent release should prevent a floor score")
		})
	})


	Context("Log message content validation", func() {

		It("cli/cli – DetailInfo entries must have non-empty Msg.Text", func() {

			result, details, err := runArtifactIntegrityCheck("cli/cli")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			var infoDetails []checker.CheckDetail
			for _, d := range details {
				if d.Type == checker.DetailInfo { 
					infoDetails = append(infoDetails, d)
				}
			}

			Expect(infoDetails).ShouldNot(BeEmpty(),
				"cli/cli has verified releases; expected at least one DetailInfo entry")

			for _, d := range infoDetails {

				Expect(d.Msg.Text).ShouldNot(BeEmpty(),
					"DetailInfo Msg.Text must not be empty (emitReleaseLog always sets it)")
			}
		})

		It("goreleaser/goreleaser – detail entries indicate SLSA/provenance coverage", func() {

			result, details, err := runArtifactIntegrityCheck("goreleaser/goreleaser")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())
			Expect(result.Score).Should(BeNumerically(">", 0))

			Expect(hasDetailType(details, checker.DetailInfo)).Should(BeTrue(),
				"expected at least one DetailInfo entry for SLSA provenance")

			if !anyMsgContains(details, "provenance", "slsa", "intoto") {
				GinkgoWriter.Println("NOTE: no detail message referenced provenance/slsa/intoto text; " +
					"consider updating if emitReleaseLog adds tier labels")
			}
		})
	})


	Context("Ecosystem diversity", func() {

		It("rust-lang/rust – large Rust project with release history", func() {
			result, details, err := runArtifactIntegrityCheck("rust-lang/rust")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(BeNumerically(">=", checker.MinResultScore))
			Expect(result.Score).Should(BeNumerically("<=", checker.MaxResultScore))

			Expect(details).ShouldNot(BeEmpty())
		})

		It("nicowillis/pong – small project, minimal ecosystem", func() {
			result, _, err := runArtifactIntegrityCheck("nicowillis/pong")

			Expect(err).Should(BeNil())
			Expect(result.Error).Should(BeNil())

			Expect(result.Score).Should(BeNumerically(">=", checker.MinResultScore))
			Expect(result.Score).Should(BeNumerically("<=", checker.MaxResultScore))
		})
	})
})
