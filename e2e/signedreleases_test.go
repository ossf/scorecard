package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
)

var _ = Describe("E2E TEST:Signedreleases", func() {
	Context("E2E TEST:Validating signed releases", func() {
		It("Should return valid signed releases", func() {
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
			result := checks.SignedReleases(checker)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
	})
})
