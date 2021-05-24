// Copyright 2021 Security Scorecard Authors
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
