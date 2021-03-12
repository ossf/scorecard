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
	"bytes"
	"fmt"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/gitcache/pkg"
)

var _ = Describe("E2E TEST:HTTP endpoint-gitcache", func() {
	var url string
	const key string = "/gitcache/github.com/ossf/scorecard"

	Context("E2E TEST:Validating http endpoint for the cache", func() {
		It("The URL to test should be available as an ENV variable", func() {
			url = os.Getenv("GITCACHE_URL")
			Expect(len(url)).ShouldNot(BeZero(), "GITCACHE_URL is not defined")
		})
		It("Should fail when an invalid git repo is passed", func() {
			jsonStr := []byte(`{"url":"http://foo/bar/scorecard"}`)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			Expect(err).Should(BeNil())

			client := &http.Client{}
			resp, err := client.Do(req)
			Expect(err).Should(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).Should(BeEquivalentTo(500))
		})
		var cache *pkg.Cache
		var err error
		It("Should be able to connect to the blob store and delete the content", func() {
			bucketURI := os.Getenv("BLOB_URL")
			Expect(len(bucketURI)).ShouldNot(BeZero(), "BLOB_URL is not defined")
			cache, err = pkg.NewBucket(bucketURI)
			Expect(err).Should(BeNil())
			Expect(cache).ShouldNot(BeNil())
			//nolint
			cache.Delete(key) // ignoring if the delete throws an error key not found
		})
		It("Liveness check should succeed", func() {
			req, err := http.NewRequest("GET", url, nil)
			Expect(err).Should(BeNil())

			client := &http.Client{}
			resp, err := client.Do(req)
			Expect(err).Should(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).Should(BeEquivalentTo(200))
		})
		It("Should be able to fetch a valid git repo multiple times", func() {
			jsonStr := []byte(`{"url":"http://github.com/ossf/scorecard"}`)
			// doing it twice to ensure the first time fetching the repository works
			// and the subsequent pulls also work.
			for i := 0; i < 2; i++ {
				req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
				Expect(err).Should(BeNil(), fmt.Sprintf("failed for %d time", i))

				client := &http.Client{}
				resp, err := client.Do(req)
				Expect(err).Should(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).Should(BeEquivalentTo(200), fmt.Sprintf("failed for %d try", i))
			}
		})
	})
})
