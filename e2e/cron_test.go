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
	"fmt"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E TEST:cron", func() {
	Context("E2E TEST:Validating cron test", func() {
		It("Should return valid test results for cron", func() {
			fileName := fmt.Sprintf("%02d-%02d-%d.json",
				time.Now().Month(), time.Now().Day(), time.Now().Year())
			_, err := ioutil.ReadFile("../" + fileName)
			Expect(err).Should(BeNil())

			// TODO - include validation to checks for file exists in the GCS bucket
		})
	})
})
