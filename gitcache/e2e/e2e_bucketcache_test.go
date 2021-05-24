// Copyright 2020 Security Scorecard Authors
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
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/gitcache/pkg"
)

var _ = Describe("E2E TEST:bucket cache", func() {
	Context("E2E TEST:Validating bucket cache", func() {
		var cache *pkg.Cache
		var err error
		It("Should be able to connect to the blob store", func() {
			bucketURI := os.Getenv("BLOB_URL")
			Expect(len(bucketURI)).ShouldNot(BeZero())
			cache, err = pkg.NewBucket(bucketURI)
			Expect(err).Should(BeNil())
			Expect(cache).ShouldNot(BeNil())
		})
		It("should be able to add content to the bucket", func() {
			const key string = "E2E_TEST_BUCKET"
			//nolint
			cache.Delete(key) // ignoring if the delete throws an error key not found
			Expect(cache.Set(key, []byte(key))).Should(BeNil())
			s, b := cache.Get(key)
			Expect(b).Should(BeTrue())
			Expect(string(s)).Should(BeEquivalentTo(key))
			Expect(cache.Delete(key)).Should(BeNil())
		})
	})
})
