// Copyright 2026 OpenSSF Scorecard Authors
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
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
)

func TestWorkflowParseErrorsIncludeFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		check fileparser.DoWhileTrueOnFileContent
		name  string
		args  []interface{}
	}{
		{
			name:  "dangerous workflows",
			check: validateGitHubActionWorkflowPatterns,
			args:  []interface{}{new(checker.DangerousWorkflowData)},
		},
		{
			name:  "token permissions",
			check: validateGitHubActionTokenPermissions,
			args:  []interface{}{new(permissionCbData)},
		},
		{
			name:  "workflow downloads",
			check: validateGitHubWorkflowIsFreeOfInsecureDownloads,
			args:  []interface{}{new(checker.PinningDependenciesData)},
		},
		{
			name:  "action pinning",
			check: validateGitHubActionWorkflow,
			args: []interface{}{
				new(checker.PinningDependenciesData),
				(*checker.CheckRequest)(nil),
				(*immutableReleaseCache)(nil),
			},
		},
		{
			name:  "SAST workflows",
			check: searchGitHubActionWorkflowUseRegex,
			args:  []interface{}{new([]string), "github/codeql-action"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			const path = ".github/workflows/broken-build.yml"
			content := []byte("on: [push\n  pull_request:\njobs:\n  build:\n    runs-on: ubuntu-latest\n")
			keepGoing, err := tt.check(path, content, tt.args...)
			if keepGoing || err == nil {
				t.Fatalf("check() = (%v, %v), want (false, parse error)", keepGoing, err)
			}
			if !strings.Contains(err.Error(), "("+path+"):") {
				t.Errorf("check() error = %q, want workflow path %q", err, path)
			}
			if !strings.Contains(err.Error(), "[syntax-check]") {
				t.Errorf("check() error = %q, want the parser diagnostic", err)
			}
		})
	}
}
