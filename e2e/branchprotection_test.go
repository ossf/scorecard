package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
)

var _ = Describe("E2E TEST:Branch Protection", func() {
	Context("E2E TEST:Validating branch protection", func() {
		It("Should fail to return branch protection on other respositories", func() {
			l := log{}
			checker := checker.Checker{
				Ctx:         context.Background(),
				Client:      ghClient,
				HttpClient:  client,
				Owner:       "apache",
				Repo:        "airflow",
				GraphClient: graphClient,
				Logf:        l.Logf,
			}
			result := checks.BranchProtection(checker)
			Expect(result.Error).ShouldNot(BeNil())
			Expect(result.Pass).Should(BeFalse())
		})
	})
})
