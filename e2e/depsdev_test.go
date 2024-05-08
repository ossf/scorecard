// Copyright 2024 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/ossf/scorecard/v5/internal/packageclient"
)

var _ = Describe("E2E TEST: depsdevclient.GetProjectPackageVersions", func() {
	var client packageclient.ProjectPackageClient

	Context("E2E TEST: Confirm ProjectPackageClient works", func() {
		It("Should receive a non-empty response from deps.dev for existing projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetProjectPackageVersions(
				context.Background(), "github.com", "ossf/scorecard",
			)
			Expect(err).Should(BeNil())
			Expect(len(versions.Versions)).Should(BeNumerically(">", 0))
		})
		It("Should error from deps.dev for nonexistent projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetProjectPackageVersions(
				context.Background(), "github.com", "ossf/scorecard-E2E-TEST-DOES-NOT-EXIST",
			)
			Expect(err).ShouldNot(BeNil())
			Expect(versions).Should(BeNil())
		})
		It("Should receive a non-empty response from deps.dev for existing projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetProjectPackageVersions(
				context.Background(), "gitlab.com", "libtiff/libtiff",
			)
			Expect(err).Should(BeNil())
			Expect(len(versions.Versions)).Should(BeNumerically(">", 0))
		})
	})
})
