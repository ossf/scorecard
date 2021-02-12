package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
)

var _ = Describe("E2E TEST:Packaging", func() {
	Context("E2E TEST:Validating use of packaging in CI/CD", func() {
		It("Should return use of packaging in CI/CD", func() {
			l := log{}
			checker := checker.Checker{
				Ctx:         context.Background(),
				Client:      ghClient,
				HttpClient:  client,
				Owner:       "apache",
				Repo:        "orc",
				GraphClient: graphClient,
				Logf:        l.Logf,
			}
			result := checks.Packaging(checker)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
		It("Should return use of packaging in CI/CD for scorecard", func() {
			l := log{}
			checker := checker.Checker{
				Ctx:         context.Background(),
				Client:      ghClient,
				HttpClient:  client,
				Owner:       "ossf",
				Repo:        "scorecard",
				GraphClient: graphClient,
				Logf:        l.Logf,
			}
			result := checks.Packaging(checker)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
	})
})
