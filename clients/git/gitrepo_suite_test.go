// Copyright 2021 OpenSSF Scorecard Authors
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

package git

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGitRepo(t *testing.T) {
	// TODO(#1709): GitHub tests are taking >10m to run slowing down CI/CD.
	// Need to fix that before re-enabling.
	// TODO(#1709): Local tests require fake Git repo to be available in CI/CD
	// environment.
	if val, exists := os.LookupEnv("RUN_GIT_E2E"); !exists || val == "0" {
		t.Skip()
	}
	if val, exists := os.LookupEnv("SKIP_GINKGO"); exists && val == "1" {
		t.Skip()
	}
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitRepo Suite")
}
