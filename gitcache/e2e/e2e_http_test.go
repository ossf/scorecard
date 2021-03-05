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
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E TEST:HTTP endpoint-gitcache", func() {
	url := "http://localhost:8080/"
	Context("E2E TEST:Validating http endpoint for the cache", func() {
		It("Should be able to fetch a valid git repo", func() {
			jsonStr := []byte(`{"url":"http://github.com/ossf/scorecard"}`)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			Expect(err).Should(BeNil())

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			Expect(resp.StatusCode).Should(BeEquivalentTo(200))
		})
		It("Should fail when an invalid git repo is passed", func() {
			jsonStr := []byte(`{"url":"http://iiiiaaa.imt/bar/scorecard"}`)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			Expect(err).Should(BeNil())

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			Expect(resp.StatusCode).Should(BeEquivalentTo(500))
		})
	})
})
