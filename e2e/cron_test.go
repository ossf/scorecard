package e2e

import (
	"context"
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/pkg"
	"github.com/ossf/scorecard/roundtripper"
)

var _ = Describe("E2E TEST:Validate cron job", func() {
	Context("E2E TEST:Validating cron job for results", func() {
		var bucket *roundtripper.Cache
		var err error
		It("Should be able to connect to the blob store", func() {
			bucketURI := os.Getenv("GCS_BUCKET")
			Expect(len(bucketURI)).ShouldNot(BeZero())
			bucket, err = roundtripper.New(context.TODO(), bucketURI)
			Expect(err).Should(BeNil())
		})
		It("Should return test results for scorecard that have valid json", func() {
			const latest string = "latest.json"
			data, b := bucket.Get(latest)
			Expect(b).Should(BeTrue())

			r := &pkg.ScorecardResults{}

			err = json.Unmarshal(data, r)
			Expect(err).Should(BeNil())
			Expect(len(r.Results)).Should(BeNumerically(">", 1))
			// cleanup the latest.json after validating
			bucket.Delete(latest)
		})
	})
})
