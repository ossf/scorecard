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

package raw

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestGithubWorkflowPinning(t *testing.T) {
	t.Parallel()

	//nolint
	tests := []struct {
		warns    int
		err      error
		name     string
		filename string
	}{
		{
			name:     "empty file",
			filename: "./testdata/.github/workflows/github-workflow-empty.yaml",
		},
		{
			name:     "comments only",
			filename: "./testdata/.github/workflows/github-workflow-comments.yaml",
		},
		{
			name:     "Pinned workflow",
			filename: "./testdata/.github/workflows/workflow-pinned.yaml",
		},
		{
			name:     "Local action workflow",
			filename: "./testdata/.github/workflows/workflow-local-action.yaml",
		},
		{
			name:     "Non-pinned workflow",
			filename: "./testdata/.github/workflows/workflow-not-pinned.yaml",
			warns:    1,
		},
		{
			name:     "Non-yaml file",
			filename: "../testdata/script.sh",
		},
		{
			name:     "Matrix as expression",
			filename: "./testdata/.github/workflows/github-workflow-matrix-expression.yaml",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			p := strings.Replace(tt.filename, "./testdata/", "", 1)
			p = strings.Replace(p, "../testdata/", "", 1)

			var r checker.PinningDependenciesData

			_, err = validateGitHubActionWorkflow(p, content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestNonGithubWorkflowPinning(t *testing.T) {
	t.Parallel()

	//nolint
	tests := []struct {
		warns    int
		err      error
		name     string
		filename string
	}{
		{
			name:     "Pinned non-github workflow",
			filename: "./testdata/.github/workflows/workflow-non-github-pinned.yaml",
		},
		{
			name:     "Pinned github workflow",
			filename: "./testdata/.github/workflows/workflow-mix-github-and-non-github-not-pinned.yaml",
			warns:    2,
		},
		{
			name:     "Pinned github workflow",
			filename: "./testdata/.github/workflows/workflow-mix-github-and-non-github-pinned.yaml",
		},
		{
			name:     "Mix of pinned and non-pinned GitHub actions",
			filename: "./testdata/.github/workflows/workflow-mix-pinned-and-non-pinned-github.yaml",
			warns:    1,
		},
		{
			name:     "Mix of pinned and non-pinned non-GitHub actions",
			filename: "./testdata/.github/workflows/workflow-mix-pinned-and-non-pinned-non-github.yaml",
			warns:    1,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
				}
			}

			p := strings.Replace(tt.filename, "./testdata/", "", 1)
			var r checker.PinningDependenciesData

			_, err = validateGitHubActionWorkflow(p, content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestGithubWorkflowPkgManagerPinning(t *testing.T) {
	t.Parallel()

	//nolint
	tests := []struct {
		warns    int
		err      error
		name     string
		filename string
	}{
		{
			name:     "npm packages without verification",
			filename: "./testdata/.github/workflows/github-workflow-pkg-managers.yaml",
			warns:    28,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			p := strings.Replace(tt.filename, "./testdata/", "", 1)
			var r checker.PinningDependenciesData

			_, err = validateGitHubWorkflowIsFreeOfInsecureDownloads(p, content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestDockerfilePinning(t *testing.T) {
	t.Parallel()

	//nolint
	tests := []struct {
		warns    int
		err      error
		name     string
		filename string
	}{
		{
			name:     "invalid dockerfile",
			filename: "./testdata/Dockerfile-invalid",
		},
		{
			name:     "invalid dockerfile sh",
			filename: "../testdata/script-sh",
		},
		{
			name:     "empty file",
			filename: "./testdata/Dockerfile-empty",
		},
		{
			name:     "comments only",
			filename: "./testdata/Dockerfile-comments",
		},
		{
			name:     "Pinned dockerfile",
			filename: "./testdata/Dockerfile-pinned",
		},
		{
			name:     "Pinned dockerfile as",
			filename: "./testdata/Dockerfile-pinned-as",
		},
		{
			name:     "Non-pinned dockerfile as",
			filename: "./testdata/Dockerfile-not-pinned-as",
			warns:    2,
		},
		{
			name:     "Non-pinned dockerfile",
			filename: "./testdata/Dockerfile-not-pinned",
			warns:    1,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
				}
			}

			var r checker.PinningDependenciesData
			_, err = validateDockerfilesPinning(tt.filename, content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestDockerfilePinningFromLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			snippet   string
			startLine uint
			endLine   uint
		}
	}{
		{
			name:     "Non-pinned dockerfile as",
			filename: "./testdata/Dockerfile-not-pinned-as",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "FROM python:3.7 as build",
					startLine: 17,
					endLine:   17,
				},
				{
					snippet:   "FROM build",
					startLine: 23,
					endLine:   23,
				},
			},
		},
		{
			name:     "Non-pinned dockerfile",
			filename: "./testdata/Dockerfile-not-pinned",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "FROM python:3.7",
					startLine: 17,
					endLine:   17,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			var r checker.PinningDependenciesData
			_, err = validateDockerfilesPinning(tt.filename, content, &r)
			if err != nil {
				t.Errorf("error during validateDockerfilesPinning: %v", err)
			}

			for _, expectedDep := range tt.expected {
				isExpectedDep := func(dep checker.Dependency) bool {
					return dep.Location.Offset == expectedDep.startLine &&
						dep.Location.EndOffset == expectedDep.endLine &&
						dep.Location.Path == tt.filename &&
						dep.Location.Snippet == expectedDep.snippet &&
						dep.Type == checker.DependencyUseTypeDockerfileContainerImage
				}

				if !scut.ValidatePinningDependencies(isExpectedDep, &r) {
					t.Errorf("test failed: dependency not present: %+v", tt.expected)
				}
			}
		})
	}
}

func TestDockerfileInvalidFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "dockerfile go",
			filename: "./testdata/Dockerfile.go",
			expected: false,
		},
		{
			name:     "dockerfile c",
			filename: "./testdata/Dockerfile.c",
			expected: false,
		},
		{
			name:     "dockerfile cpp",
			filename: "./testdata/Dockerfile.cpp",
			expected: false,
		},
		{
			name:     "dockerfile rust",
			filename: "./testdata/Dockerfile.rs",
			expected: false,
		},
		{
			name:     "dockerfile js",
			filename: "./testdata/Dockerfile.js",
			expected: false,
		},
		{
			name:     "dockerfile sh",
			filename: "./testdata/Dockerfile.sh",
			expected: false,
		},
		{
			name:     "dockerfile py",
			filename: "./testdata/Dockerfile.py",
			expected: false,
		},
		{
			name:     "dockerfile pyc",
			filename: "./testdata/Dockerfile.pyc",
			expected: false,
		},
		{
			name:     "dockerfile java",
			filename: "./testdata/Dockerfile.java",
			expected: false,
		},
		{
			name:     "dockerfile ",
			filename: "./testdata/Dockerfile.any",
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var c []byte
			r := isDockerfile(tt.filename, c)
			if r != tt.expected {
				t.Errorf("test failed: %s. Expected %v. Got %v", tt.filename, r, tt.expected)
			}
		})
	}
}

func TestDockerfileInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		filename string
		expected []struct {
			snippet   string
			startLine uint
			endLine   uint
			t         checker.DependencyUseType
		}
	}{
		{
			name:     "dockerfile downloads",
			filename: "./testdata/Dockerfile-download-lines",
			//nolint
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
				t         checker.DependencyUseType
			}{
				{
					snippet:   "curl bla | bash",
					startLine: 35,
					endLine:   36,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "pip install -r requirements.txt",
					startLine: 41,
					endLine:   42,
					t:         checker.DependencyUseTypePipCommand,
				},
			},
		},
		{
			name:     "dockerfile downloads multi-run",
			filename: "./testdata/Dockerfile-download-multi-runs",
			//nolint
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
				t         checker.DependencyUseType
			}{
				{
					snippet:   "/tmp/file3",
					startLine: 28,
					endLine:   28,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "/tmp/file1",
					startLine: 30,
					endLine:   30,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "bash /tmp/file3",
					startLine: 32,
					endLine:   34,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "bash /tmp/file1",
					startLine: 37,
					endLine:   38,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			var r checker.PinningDependenciesData
			_, err = validateDockerfileInsecureDownloads(tt.filename, content, &r)
			if err != nil {
				t.Errorf("error during validateDockerfileInsecureDownloads: %v", err)
			}

			for _, expectedDep := range tt.expected {
				isExpectedDep := func(dep checker.Dependency) bool {
					return dep.Location.Offset == expectedDep.startLine &&
						dep.Location.EndOffset == expectedDep.endLine &&
						dep.Location.Path == tt.filename &&
						dep.Location.Snippet == expectedDep.snippet &&
						dep.Type == expectedDep.t
				}

				if !scut.ValidatePinningDependencies(isExpectedDep, &r) {
					t.Errorf("test failed: dependency not present: %+v", tt.expected)
				}
			}
		})
	}
}

func TestShellscriptInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		filename string
		expected []struct {
			snippet   string
			startLine uint
			endLine   uint
			t         checker.DependencyUseType
		}
	}{
		{
			name:     "shell downloads",
			filename: "./testdata/shell-download-lines.sh",
			//nolint
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
				t         checker.DependencyUseType
			}{
				{
					snippet:   "bash /tmp/file",
					startLine: 6,
					endLine:   6,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 11,
					endLine:   11,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "bash <(wget -qO- http://website.com/my-script.sh)",
					startLine: 18,
					endLine:   18,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "bash <(wget -qO- http://website.com/my-script.sh)",
					startLine: 20,
					endLine:   20,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "pip install -r requirements.txt",
					startLine: 26,
					endLine:   26,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 28,
					endLine:   28,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "choco install 'some-package'",
					startLine: 30,
					endLine:   30,
					t:         checker.DependencyUseTypeChocoCommand,
				},
				{
					snippet:   "choco install 'some-other-package'",
					startLine: 31,
					endLine:   31,
					t:         checker.DependencyUseTypeChocoCommand,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			var r checker.PinningDependenciesData
			_, err = validateShellScriptIsFreeOfInsecureDownloads(tt.filename, content, &r)
			if err != nil {
				t.Errorf("error during validateShellScriptIsFreeOfInsecureDownloads: %v", err)
			}

			for _, expectedDep := range tt.expected {
				isExpectedDep := func(dep checker.Dependency) bool {
					return dep.Location.Offset == expectedDep.startLine &&
						dep.Location.EndOffset == expectedDep.endLine &&
						dep.Location.Path == tt.filename &&
						dep.Type == expectedDep.t &&
						dep.Location.Snippet == expectedDep.snippet
				}

				if !scut.ValidatePinningDependencies(isExpectedDep, &r) {
					t.Errorf("test failed: dependency not present: %+v", tt.expected)
				}
			}
		})
	}
}

func TestDockerfilePinningWihoutHash(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		warns    int
		err      error
		name     string
		filename string
	}{
		{
			name:     "Pinned dockerfile as no hash",
			filename: "./testdata/Dockerfile-pinned-as-without-hash",
			warns:    4,
		},
		{
			name:     "Dockerfile with args",
			filename: "./testdata/Dockerfile-args",
			warns:    2,
		},
		{
			name:     "Dockerfile with base",
			filename: "./testdata/Dockerfile-base",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			var r checker.PinningDependenciesData
			_, err = validateDockerfilesPinning(tt.filename, content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestDockerfileScriptDownload(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		warns    int
		err      error
		name     string
		filename string
	}{
		{
			name:     "curl | sh",
			filename: "./testdata/Dockerfile-curl-sh",
			warns:    4,
		},
		{
			name:     "empty file",
			filename: "./testdata/Dockerfile-empty",
		},
		{
			name:     "invalid file sh",
			filename: "../testdata/script.sh",
		},
		{
			name:     "comments only",
			filename: "./testdata/Dockerfile-comments",
		},
		{
			name:     "wget | /bin/sh",
			filename: "./testdata/Dockerfile-wget-bin-sh",
			warns:    3,
		},
		{
			name:     "wget no exec",
			filename: "./testdata/Dockerfile-script-ok",
		},
		{
			name:     "curl file sh",
			filename: "./testdata/Dockerfile-curl-file-sh",
			warns:    12,
		},
		{
			name:     "proc substitution",
			filename: "./testdata/Dockerfile-proc-subs",
			warns:    6,
		},
		{
			name:     "wget file",
			filename: "./testdata/Dockerfile-wget-file",
			warns:    10,
		},
		{
			name:     "gsutil file",
			filename: "./testdata/Dockerfile-gsutil-file",
			warns:    17,
		},
		{
			name:     "aws file",
			filename: "./testdata/Dockerfile-aws-file",
			warns:    15,
		},
		{
			name:     "pkg managers",
			filename: "./testdata/Dockerfile-pkg-managers",
			warns:    39,
		},
		{
			name:     "download with some python",
			filename: "./testdata/Dockerfile-some-python",
			warns:    1,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
				}
			}

			var r checker.PinningDependenciesData
			_, err = validateDockerfileInsecureDownloads(tt.filename, content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestDockerfileScriptDownloadInfo(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		filename string
		warns    int
		err      error
	}{
		{
			name:     "curl | sh",
			filename: "./testdata/Dockerfile-no-curl-sh",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}
			var r checker.PinningDependenciesData
			_, err = validateDockerfileInsecureDownloads(tt.filename, content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestShellScriptDownload(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		filename string
		warns    int
		debugs   int
		err      error
	}{
		{
			name:     "sh script",
			filename: "../testdata/script-sh",
			warns:    7,
		},
		{
			name:     "empty file",
			filename: "./testdata/script-empty.sh",
		},
		{
			name:     "comments",
			filename: "./testdata/script-comments.sh",
		},
		{
			name:     "bash script",
			filename: "./testdata/script-bash",
			warns:    7,
		},
		{
			name:     "sh script 2",
			filename: "../testdata/script.sh",
			warns:    7,
		},
		{
			name:     "pkg managers",
			filename: "./testdata/script-pkg-managers",
			warns:    36,
		},
		{
			name:     "invalid shell script",
			filename: "./testdata/script-invalid.sh",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
				}
			}

			var r checker.PinningDependenciesData
			_, err = validateShellScriptIsFreeOfInsecureDownloads(tt.filename, content, &r)

			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			// Note: this works because all our examples
			// either have warns or debugs.
			ws := (tt.warns == len(r.Dependencies)) && (tt.debugs == 0)
			ds := (tt.debugs == len(r.Dependencies)) && (tt.warns == 0)
			if !ws && !ds {
				t.Errorf("expected %v or %v. Got %v", tt.warns, tt.debugs, len(r.Dependencies))
			}
		})
	}
}

func TestShellScriptDownloadPinned(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		filename string
		warns    int
		err      error
	}{
		{
			name:     "sh script",
			filename: "./testdata/script-comments.sh",
		},
		{
			name:     "script free of download",
			filename: "./testdata/script-free-from-download.sh",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			var r checker.PinningDependenciesData
			_, err = validateShellScriptIsFreeOfInsecureDownloads(tt.filename, content, &r)

			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestGitHubWorflowRunDownload(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		filename string
		warns    int
		err      error
	}{
		{
			name:     "workflow curl default",
			filename: "./testdata/.github/workflows/github-workflow-curl-default.yaml",
			warns:    1,
		},
		{
			name:     "workflow curl no default",
			filename: "./testdata/.github/workflows/github-workflow-curl-no-default.yaml",
			warns:    1,
		},
		{
			name:     "wget across steps",
			filename: "./testdata/.github/workflows/github-workflow-wget-across-steps.yaml",
			warns:    2,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
				}
			}
			p := strings.Replace(tt.filename, "./testdata/", "", 1)

			var r checker.PinningDependenciesData

			_, err = validateGitHubWorkflowIsFreeOfInsecureDownloads(p, content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			if tt.warns != len(r.Dependencies) {
				t.Errorf("expected %v. Got %v", tt.warns, len(r.Dependencies))
			}
		})
	}
}

func TestGitHubWorkflowUsesLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			dependency string
			startLine  uint
			endLine    uint
		}
	}{
		{
			name:     "unpinned dependency in uses",
			filename: "../testdata/.github/workflows/github-workflow-permissions-run-codeql-write.yaml",
			expected: []struct {
				dependency string
				startLine  uint
				endLine    uint
			}{
				{
					dependency: "github/codeql-action/analyze@v1",
					startLine:  25,
					endLine:    25,
				},
			},
		},
		{
			name:     "multiple unpinned dependency in uses",
			filename: "./testdata/.github/workflows/github-workflow-multiple-unpinned-uses.yaml",
			expected: []struct {
				dependency string
				startLine  uint
				endLine    uint
			}{
				{
					dependency: "github/codeql-action/analyze@v1",
					startLine:  22,
					endLine:    22,
				},
				{
					dependency: "docker/build-push-action@1.2.3",
					startLine:  24,
					endLine:    24,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			p := strings.Replace(tt.filename, "../testdata/", "", 1)
			p = strings.Replace(p, "./testdata/", "", 1)
			var r checker.PinningDependenciesData

			_, err = validateGitHubActionWorkflow(p, content, &r)
			if err != nil {
				t.Errorf("validateGitHubActionWorkflow: %v", err)
			}
			for _, expectedDep := range tt.expected {
				isExpectedDep := func(dep checker.Dependency) bool {
					return dep.Location.Offset == expectedDep.startLine &&
						dep.Location.EndOffset == expectedDep.endLine &&
						dep.Location.Path == p &&
						dep.Location.Snippet == expectedDep.dependency &&
						dep.Type == checker.DependencyUseTypeGHAction
				}

				if !scut.ValidatePinningDependencies(isExpectedDep, &r) {
					t.Errorf("test failed: dependency not present: %+v", tt.expected)
				}
			}
		})
	}
}

func TestGitHubWorkInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			snippet   string
			startLine uint
			endLine   uint
		}
	}{
		{
			name:     "downloads",
			filename: "./testdata/.github/workflows/github-workflow-download-lines.yaml",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "bash /tmp/file",
					startLine: 27,
					endLine:   27,
				},
				{
					snippet:   "/tmp/file2",
					startLine: 29,
					endLine:   29,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 32,
					endLine:   32,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			p := strings.Replace(tt.filename, "./testdata/", "", 1)
			var r checker.PinningDependenciesData

			_, err = validateGitHubWorkflowIsFreeOfInsecureDownloads(p, content, &r)
			if err != nil {
				t.Errorf("error during validateGitHubWorkflowIsFreeOfInsecureDownloads: %v", err)
			}

			for _, expectedDep := range tt.expected {
				isExpectedDep := func(dep checker.Dependency) bool {
					return dep.Location.Offset == expectedDep.startLine &&
						dep.Location.EndOffset == expectedDep.endLine &&
						dep.Location.Path == p &&
						dep.Location.Snippet == expectedDep.snippet &&
						dep.Type == checker.DependencyUseTypeDownloadThenRun
				}

				if !scut.ValidatePinningDependencies(isExpectedDep, &r) {
					t.Errorf("test failed: dependency not present: %+v", tt.expected)
				}
			}
		})
	}
}
