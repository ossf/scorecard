package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
)

var _ = Describe("E2E TEST:Fuzzing", func() {
	Context("E2E TEST:Validating use of fuzzing tools", func() {
		It("Should return use of fuzzing tools", func() {
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
			result := checks.Fuzzing(checker)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
	})
})
