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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/checks"
	"github.com/ossf/scorecard/checker"
)

var _ = Describe("E2E TEST:Packaging", func() {
	Context("E2E TEST:Validating use of packaging in CI/CD", func() {
		It("Should return use of packaging in CI/CD", func() {
			l := log{}
			checkRequest := checker.CheckRequest{
				Ctx:         context.Background(),
				Client:      ghClient,
				HttpClient:  client,
				Owner:       "apache",
				Repo:        "orc",
				GraphClient: graphClient,
				Logf:        l.Logf,
			}
			result := checks.Packaging(checkRequest)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
		It("Should return use of packaging in CI/CD for scorecard", func() {
			l := log{}
			lib := checker.CheckRequest{
				Ctx:         context.Background(),
				Client:      ghClient,
				HttpClient:  client,
				Owner:       "ossf",
				Repo:        "scorecard",
				GraphClient: graphClient,
				Logf:        l.Logf,
			}
			result := checks.Packaging(lib)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
	})
})
