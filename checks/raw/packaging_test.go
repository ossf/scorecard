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

package raw

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/rhysd/actionlint"
)

func TestIsPackagingWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "npmjs.org publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-npm.yaml",
			expected: true,
		},
		{
			name:     "npm github publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-npm-github.yaml",
			expected: false, // Should this be false?
		},
		{
			name:     "maven publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-maven.yaml",
			expected: true,
		},
		{
			name:     "gradle publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-gradle.yaml",
			expected: true,
		},
		{
			name:     "gem publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-gem.yaml",
			expected: true,
		},
		{
			name:     "nuget publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-nuget.yaml",
			expected: true,
		},
		{
			name:     "docker action publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-docker-action.yaml",
			expected: true,
		},
		{
			name:     "docker push publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-docker-push.yaml",
			expected: true,
		},
		{
			name:     "pypi publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-pypi.yaml",
			expected: true,
		},
		{
			name:     "python semantic release publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-python-semantic-release.yaml",
			expected: true,
		},
		{
			name:     "go publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-go.yaml",
			expected: true,
		},
		{
			name:     "cargo publish",
			filename: "./testdata/.github/workflows/github-workflow-packaging-cargo.yaml",
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				panic(fmt.Errorf("cannot read file: %w", err))
			}
			workflow, errs := actionlint.Parse(content)
			if len(errs) > 0 && workflow == nil {
				panic(fmt.Errorf("cannot parse file: %w", err))
			}
			p := strings.Replace(tt.filename, "./testdata/", "", 1)

			_, ok := isPackagingWorkflow(workflow, p)
			if ok != tt.expected {
				t.Errorf("isPackagingWorkflow() = %v, expected %v", ok, tt.expected)
			}
		})
	}
}
