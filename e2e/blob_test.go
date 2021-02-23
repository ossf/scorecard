package e2e

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/roundtripper"
)

var _ = Describe("E2E TEST:blob", func() {
	Context("E2E TEST:Validating blob test", func() {
		It("Should be able to connect to the blob store", func() {
			bucketURI := os.Getenv("BLOB_URL")
			Expect(len(bucketURI)).ShouldNot(BeZero())
			bucket, err := roundtripper.New(context.TODO(), bucketURI)
			Expect(err).Should(BeNil())

			const key string = "ossf-scorecard-test"
			const value string = "ossf-scorecard-test"
			bucket.Set(key, []byte(value))
			defer bucket.Delete(key)

			v, b := bucket.Get(key)
			Expect(b).Should(BeTrue())
			Expect(string(v)).Should(BeEquivalentTo(value))
		})
	})
})
