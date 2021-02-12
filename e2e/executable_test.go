package e2e

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type scorecard struct {
	Repo   string `json:"Repo"`
	Date   string `json:"Date"`
	Checks []struct {
		CheckName  string   `json:"CheckName"`
		Pass       bool     `json:"Pass"`
		Confidence int      `json:"Confidence"`
		Details    []string `json:"Details"`
	} `json:"Checks"`
}

var _ = Describe("E2E TEST:executable", func() {
	Context("E2E TEST:Validating executable test", func() {
		It("Should return valid test results for scorecard", func() {

			file, _ := ioutil.ReadFile("../bin/results.json")

			data := scorecard{}

			err := json.Unmarshal([]byte(file), &data)

			Expect(err).Should(BeNil())

			for _, c := range data.Checks {
				switch c.CheckName {
				case "Active":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "Branch-Protection":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "CI-Tests":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "CII-Best-Practices":
					Expect(c.Pass).Should(BeFalse(), c.CheckName)
				case "Code-Review":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "Contributors":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "Frozen-Deps":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "Fuzzing":
					Expect(c.Pass).Should(BeFalse(), c.CheckName)
				case "Packaging":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "Pull-Requests":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "SAST":
					Expect(c.Pass).Should(BeTrue(), c.CheckName)
				case "Signed-Releases":
					Expect(c.Pass).Should(BeFalse(), c.CheckName)
				case "Signed-Tags":
					Expect(c.Pass).Should(BeFalse(), c.CheckName)
				}
			}
		})
	})
})
