// Copyright 2024 OpenSSF Scorecard Authors
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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v5/checks"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/pkg"
)

var _ = Describe("E2E TEST: config parsing", func() {
	Context("E2E TEST:Valid config parsing", func() {
		It("Should return an annotation from the config", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-config-e2e")
			Expect(err).Should(BeNil())
			results, err := pkg.Run(context.Background(), repo,
				pkg.WithChecks([]string{checks.CheckCodeReview}),
			)
			Expect(err).Should(BeNil())
			Expect(len(results.Config.Annotations)).Should(BeNumerically(">=", 1))
		})
	})
})
