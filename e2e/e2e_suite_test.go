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
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/ossf/scorecard/v4/clients/githubrepo"
)

var logger *zap.Logger

func TestE2e(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	// making sure the GITHUB_AUTH_TOKEN is set prior to running e2e tests
	token, contains := os.LookupEnv("GITHUB_AUTH_TOKEN")

	Expect(contains).ShouldNot(BeFalse(),
		"GITHUB_AUTH_TOKEN env variable is not set.The GITHUB_AUTH_TOKEN env variable has to be set to run e2e test.")
	Expect(len(token)).ShouldNot(BeZero(), "Length of the GITHUB_AUTH_TOKEN env variable is zero.")

	l, err := githubrepo.NewLogger(zap.InfoLevel)
	Expect(err).Should(BeNil())
	logger = l
})

var _ = AfterSuite(func() {
})
