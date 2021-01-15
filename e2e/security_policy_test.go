package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
)

var _ = Describe("E2E TEST:SecurityPolicy", func() {
	Context("E2E TEST:Validating security policy", func() {
		It("Should return valid security policy", func() {
			l := log{}
			checker := checker.Checker{
				Ctx:         context.Background(),
				Client:      ghClient,
				HttpClient:  client,
				Owner:       "tensorflow",
				Repo:        "tensorflow",
				GraphClient: graphClient,
				Logf:        l.Logf,
			}
			result := checks.SecurityPolicy(checker)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
	})
})
