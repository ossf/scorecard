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

	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
)

var _ = Describe("E2E TEST:Automatic-Dependency-Update", func() {
	Context("E2E TEST:Validating dependencies are automatically updated", func() {
		It("Should return deps are automatically updated for dependabot", func() {
			l := log{}
			checker := checker.CheckRequest{
				Ctx:         context.Background(),
				Client:      ghClient,
				Owner:       "ossf",
				Repo:        "scorecard",
				GraphClient: graphClient,
				Logf:        l.Logf,
			}
			result := checks.AutomaticDependencyUpdate(&checker)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
		It("Should return deps are automatically updated for renovatebot", func() {
			l := log{}
			checker := checker.CheckRequest{
				Ctx:         context.Background(),
				Client:      ghClient,
				Owner:       "netlify",
				Repo:        "netlify-cms",
				GraphClient: graphClient,
				Logf:        l.Logf,
			}
			result := checks.AutomaticDependencyUpdate(&checker)
			Expect(result.Error).Should(BeNil())
			Expect(result.Pass).Should(BeTrue())
		})
	})
})
