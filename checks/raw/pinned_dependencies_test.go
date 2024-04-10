// Copyright 2020 OpenSSF Scorecard Authors
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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestGithubWorkflowPinning(t *testing.T) {
	t.Parallel()

	//nolint:govet
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
			warns:    2,
		},
		{
			name:     "Non-yaml file",
			filename: "../testdata/script.sh",
		},
		{
			name:     "Matrix as expression",
			filename: "./testdata/.github/workflows/github-workflow-matrix-expression.yaml",
		},
		{
			name:     "Can't detect OS, but still detects unpinned Actions",
			filename: "./testdata/.github/workflows/github-workflow-unknown-os.yaml",
			warns:    2, // 1 in job with unknown OS, 1 in job with known OS
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.warns != unpinned {
				t.Errorf("expected %v. Got %v", tt.warns, unpinned)
			}
		})
	}
}

func TestGithubWorkflowPinningPattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		uses     string
		ispinned bool
	}{
		{
			desc:     "checking out mutable tag",
			uses:     "actions/checkout@v3",
			ispinned: false,
		},
		{
			desc:     "checking out mutable tag",
			uses:     "actions/checkout@v3.2.0",
			ispinned: false,
		},
		{
			desc:     "checking out mutable tag",
			uses:     "actions/checkout@main",
			ispinned: false,
		},
		{
			desc:     "checking out mutable tag",
			uses:     "actions/aws@v2.0.1",
			ispinned: false,
		},
		{
			desc:     "checking out mutable tag",
			uses:     "actions/aws/ec2@main",
			ispinned: false,
		},
		{
			desc:     "checking out specific commit from github with truncated SHA-1",
			uses:     "actions/checkout@a81bbbf",
			ispinned: false,
		},
		{
			desc:     "checking out specific commit from github with SHA-1",
			uses:     "actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675",
			ispinned: true,
		},
		{
			desc:     "local workflow",
			uses:     "./.github/uses.yml",
			ispinned: true,
		},
		{
			desc:     "non-github docker image pinned by digest",
			uses:     "docker://gcr.io/distroless/static-debian11@sha256:9e6f8952f12974d088f648ed6252ea1887cdd8641719c8acd36bf6d2537e71c0",
			ispinned: true,
		},
		{
			desc:     "non-github docker image pinned to mutable tag",
			uses:     "docker://gcr.io/distroless/static-debian11:sha256-3876708467ad6f38f263774aa107d331e8de6558a2874aa223b96fc0d9dfc820.sig",
			ispinned: false,
		},
		{
			desc:     "non-github docker image pinned to mutable version",
			uses:     "docker://rhysd/actionlint:latest",
			ispinned: false,
		},
		{
			desc:     "non-github docker image pinned by digest",
			uses:     "docker://rhysd/actionlint:latest@sha256:5f957b2a08d223e48133e1a914ed046bea12e578fe2f6ae4de47fdbe691a2468",
			ispinned: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()
			p := isActionDependencyPinned(tt.uses)
			if p != tt.ispinned {
				t.Fatalf("dependency %s ispinned?: %v expected?: %v", tt.uses, p, tt.ispinned)
			}
		})
	}
}

func TestNonGithubWorkflowPinning(t *testing.T) {
	t.Parallel()

	//nolint:govet
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
		{
			name:     "Can't detect OS, but still detects unpinned Actions",
			filename: "./testdata/.github/workflows/github-workflow-unknown-os.yaml",
			warns:    2, // 1 in job with unknown OS, 1 in job with known OS
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.warns != unpinned {
				t.Errorf("expected %v. Got %v", tt.warns, unpinned)
			}
		})
	}
}

func TestGithubWorkflowPkgManagerPinning(t *testing.T) {
	t.Parallel()

	//nolint:govet
	tests := []struct {
		unpinned         int
		processingErrors int
		err              error
		name             string
		filename         string
	}{
		{
			name:     "npm packages without verification",
			filename: "./testdata/.github/workflows/github-workflow-pkg-managers.yaml",
			unpinned: 49,
		},
		{
			name:             "Can't identify OS but doesn't crash",
			filename:         "./testdata/.github/workflows/github-workflow-unknown-os.yaml",
			processingErrors: 1, // job with unknown OS is skipped
			unpinned:         1, // only 1 in job with known OS, since other job is skipped
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.unpinned != unpinned {
				t.Errorf("expected %v unpinned. Got %v", tt.unpinned, unpinned)
			}

			if tt.processingErrors != len(r.ProcessingErrors) {
				t.Errorf("expected %v processing errors. Got %v", tt.processingErrors, len(r.ProcessingErrors))
			}
		})
	}
}

func TestDockerfilePinning(t *testing.T) {
	t.Parallel()

	//nolint:govet
	tests := []struct {
		warns    int
		err      error
		name     string
		filename string
	}{
		{
			name:     "invalid dockerfile",
			filename: "Dockerfile-invalid",
		},
		{
			name:     "invalid dockerfile sh",
			filename: "../../testdata/script-sh",
		},
		{
			name:     "empty file",
			filename: "Dockerfile-empty",
		},
		{
			name:     "comments only",
			filename: "Dockerfile-comments",
		},
		{
			name:     "Pinned dockerfile",
			filename: "Dockerfile-pinned",
		},
		{
			name:     "Pinned dockerfile as",
			filename: "Dockerfile-pinned-as",
		},
		{
			name:     "Non-pinned dockerfile as",
			filename: "Dockerfile-not-pinned-as",
			warns:    2,
		},
		{
			name:     "Non-pinned dockerfile but in vendor, ie: 0 warns",
			filename: "vendor/Dockerfile-not-pinned-as",
			warns:    0,
		},
		{
			name:     "Non-pinned dockerfile",
			filename: "Dockerfile-not-pinned",
			warns:    1,
		},
		{
			name:     "Parser error doesn't affect docker image pinning",
			filename: "Dockerfile-not-pinned-with-parser-error",
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
				content, err = os.ReadFile(filepath.Join("testdata", tt.filename))
				if err != nil {
					t.Errorf("cannot read file: %v", err)
				}
			}

			var r checker.PinningDependenciesData
			_, err = validateDockerfilesPinning(filepath.Join("testdata", tt.filename), content, &r)
			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			unpinned := countUnpinned(r.Dependencies)

			if tt.warns != unpinned {
				t.Errorf("expected %v. Got %v", tt.warns, unpinned)
			}
		})
	}
}

func TestFileIsInVendorDir(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "not in vendor or third_party",
			filename: "a/b/c/d/Dockerfile",
			expected: false,
		},
		{
			name:     "is third_party deep in tree",
			filename: "a/b/third_party/Dockerfile",
			expected: true,
		},
		{
			name:     "in vendor",
			filename: "vendor/a/b/Dockerfile",
			expected: true,
		},
		{
			name:     "in third_party",
			filename: "third_party/b/c/Dockerfile",
			expected: true,
		},
		{
			name:     "in deep vendor",
			filename: "a/b/c/vendor/Dockerfile",
			expected: true,
		},
		{
			name:     "misspelled vendor dir",
			filename: "a/vendor_/Dockerfile",
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fileIsInVendorDir(tt.filename)
			if got != tt.expected {
				t.Errorf("expected %v. Got %v", tt.expected, got)
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
		{
			name:     "Parser error doesn't affect docker image pinning",
			filename: "./testdata/Dockerfile-not-pinned-with-parser-error",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "FROM abrarov/msvc-2017:2.11.0",
					startLine: 1,
					endLine:   1,
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

func TestDockerfileInsecureDownloadsBrokenCommands(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		filename string
		err      error
	}{
		{
			name:     "dockerfile downloads",
			filename: "./testdata/Dockerfile-empty-run-array",
			err:      errInternalInvalidDockerFile,
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
			if !strings.Contains(err.Error(), tt.err.Error()) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}
		})
	}
}

func TestDockerfileInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name             string
		filename         string
		processingErrors int
		expected         []struct {
			snippet   string
			startLine uint
			endLine   uint
			t         checker.DependencyUseType
		}
	}{
		{
			name:     "dockerfile downloads",
			filename: "./testdata/Dockerfile-download-lines",
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
				{
					snippet:   "pip install --no-deps -e hg+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package",
					startLine: 46,
					endLine:   46,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e svn+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package",
					startLine: 47,
					endLine:   47,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e bzr+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package",
					startLine: 48,
					endLine:   48,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git",
					startLine: 49,
					endLine:   49,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git#egg=package",
					startLine: 50,
					endLine:   50,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@v1.0",
					startLine: 51,
					endLine:   51,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@v1.0#egg=package",
					startLine: 52,
					endLine:   52,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package",
					startLine: 60,
					endLine:   60,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e . git+https://github.com/username/repo.git",
					startLine: 61,
					endLine:   61,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "python -m pip install --no-deps -e git+https://github.com/username/repo.git",
					startLine: 64,
					endLine:   64,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   `bash <(curl --silent --show-error "https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash")`,
					startLine: 68,
					endLine:   68,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "curl -sSL https://dot.net/v1/dotnet-install.sh | bash /dev/stdin",
					startLine: 69,
					endLine:   69,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "curl -sSL https://raw.githubusercontent.com/dotnet/install-scripts/main/src/dotnet-install.sh | bash /dev/stdin",
					startLine: 70,
					endLine:   70,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
			},
		},
		{
			name:     "dockerfile downloads multi-run",
			filename: "./testdata/Dockerfile-download-multi-runs",
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
		{
			name:             "Parser error may lead to incomplete data",
			filename:         "./testdata/Dockerfile-not-pinned-with-parser-error",
			processingErrors: 1,
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
				t         checker.DependencyUseType
			}{
				{
					snippet:   "choco install --no-progress -r -y cmake",
					startLine: 4,
					endLine:   4,
					t:         checker.DependencyUseTypeChocoCommand,
				},
				{
					snippet:   "choco install --no-progress -r -y gzip wget ninja",
					startLine: 9,
					endLine:   9,
					t:         checker.DependencyUseTypeChocoCommand,
				},
				// `curl bla | bash` isn't detected due to parser error
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

			if tt.processingErrors != len(r.ProcessingErrors) {
				t.Errorf("expected %v processing errors. Got %v", tt.processingErrors, len(r.ProcessingErrors))
			}
		})
	}
}

func TestDockerfileWithHeredocsInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name             string
		filename         string
		processingErrors int
		expected         []struct {
			snippet   string
			startLine uint
			endLine   uint
			pinned    bool
			t         checker.DependencyUseType
		}
	}{
		{
			name:             "dockerfile heredoc downloads",
			filename:         "./testdata/Dockerfile-download-heredoc",
			processingErrors: 1,
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
				pinned    bool
				t         checker.DependencyUseType
			}{
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@v1.0#egg=package",
					startLine: 20,
					endLine:   20,
					pinned:    false,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567",
					startLine: 24,
					endLine:   24,
					pinned:    true,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 28,
					endLine:   28,
					pinned:    false,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567",
					startLine: 32,
					endLine:   32,
					pinned:    true,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@v1.0#egg=package",
					startLine: 36,
					endLine:   36,
					pinned:    false,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 38,
					endLine:   38,
					pinned:    false,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567",
					startLine: 42,
					endLine:   43,
					pinned:    true,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@v1.0#egg=package",
					startLine: 43,
					endLine:   44,
					pinned:    false,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 45,
					endLine:   45,
					pinned:    false,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567",
					startLine: 50,
					endLine:   52,
					pinned:    true,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 51,
					endLine:   53,
					pinned:    false,
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
						*dep.Pinned == expectedDep.pinned &&
						dep.Type == expectedDep.t
				}

				if !scut.ValidatePinningDependencies(isExpectedDep, &r) {
					t.Errorf("test failed: dependency not present: %+v", tt.expected)
				}
			}

			if tt.processingErrors != len(r.ProcessingErrors) {
				t.Errorf("expected %v processing errors. Got %v", tt.processingErrors, len(r.ProcessingErrors))
			}
		})
	}
}

func TestShellscriptInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	//nolint:govet
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
				{
					snippet:   "pip install --no-deps -e hg+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package",
					startLine: 38,
					endLine:   38,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e svn+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package",
					startLine: 39,
					endLine:   39,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e bzr+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package",
					startLine: 40,
					endLine:   40,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git",
					startLine: 41,
					endLine:   41,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git#egg=package",
					startLine: 42,
					endLine:   42,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@v1.0",
					startLine: 43,
					endLine:   43,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e git+https://github.com/username/repo.git@v1.0#egg=package",
					startLine: 44,
					endLine:   44,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package",
					startLine: 52,
					endLine:   52,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "pip install --no-deps -e . git+https://github.com/username/repo.git",
					startLine: 53,
					endLine:   53,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "python -m pip install --no-deps -e git+https://github.com/username/repo.git",
					startLine: 56,
					endLine:   56,
					t:         checker.DependencyUseTypePipCommand,
				},
				{
					snippet:   "nuget install some-package",
					startLine: 59,
					endLine:   59,
					t:         checker.DependencyUseTypeNugetCommand,
				},
				{
					snippet:   "dotnet add package some-package",
					startLine: 63,
					endLine:   63,
					t:         checker.DependencyUseTypeNugetCommand,
				},
				{
					snippet:   "dotnet add SomeProject package some-package",
					startLine: 64,
					endLine:   64,
					t:         checker.DependencyUseTypeNugetCommand,
				},
				{
					snippet:   `bash <(curl --silent --show-error "https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash")`,
					startLine: 69,
					endLine:   69,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "curl -sSL https://dot.net/v1/dotnet-install.sh | bash /dev/stdin",
					startLine: 70,
					endLine:   70,
					t:         checker.DependencyUseTypeDownloadThenRun,
				},
				{
					snippet:   "curl -sSL https://raw.githubusercontent.com/dotnet/install-scripts/main/src/dotnet-install.sh | bash /dev/stdin",
					startLine: 71,
					endLine:   71,
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

func TestDockerfilePinningWithoutHash(t *testing.T) {
	t.Parallel()
	//nolint:govet
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.warns != unpinned {
				t.Errorf("expected %v. Got %v", tt.warns, unpinned)
			}
		})
	}
}

func TestDockerfileScriptDownload(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		unpinned         int
		processingErrors int
		err              error
		name             string
		filename         string
	}{
		{
			name:     "curl | sh",
			filename: "./testdata/Dockerfile-curl-sh",
			unpinned: 5,
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
			unpinned: 4,
		},
		{
			name:     "wget no exec",
			filename: "./testdata/Dockerfile-script-ok",
		},
		{
			name:     "curl file sh",
			filename: "./testdata/Dockerfile-curl-file-sh",
			unpinned: 12,
		},
		{
			name:     "proc substitution",
			filename: "./testdata/Dockerfile-proc-subs",
			unpinned: 6,
		},
		{
			name:     "wget file",
			filename: "./testdata/Dockerfile-wget-file",
			unpinned: 10,
		},
		{
			name:     "gsutil file",
			filename: "./testdata/Dockerfile-gsutil-file",
			unpinned: 17,
		},
		{
			name:     "aws file",
			filename: "./testdata/Dockerfile-aws-file",
			unpinned: 15,
		},
		{
			name:     "pkg managers",
			filename: "./testdata/Dockerfile-pkg-managers",
			unpinned: 60,
		},
		{
			name:     "download with some python",
			filename: "./testdata/Dockerfile-some-python",
			unpinned: 1,
		},
		{
			name:             "Parser error doesn't affect docker image pinning",
			filename:         "./testdata/Dockerfile-not-pinned-with-parser-error",
			processingErrors: 1,
			unpinned:         2, // `curl bla | bash` missed due to parser error
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.unpinned != unpinned {
				t.Errorf("expected %v unpinned. Got %v", tt.unpinned, unpinned)
			}

			if tt.processingErrors != len(r.ProcessingErrors) {
				t.Errorf("expected %v processing errors. Got %v", tt.processingErrors, len(r.ProcessingErrors))
			}
		})
	}
}

func TestDockerfileScriptDownloadInfo(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name             string
		filename         string
		unpinned         int
		processingErrors int
		err              error
	}{
		{
			name:     "curl | sh",
			filename: "./testdata/Dockerfile-no-curl-sh",
		},
		{
			name:             "Parser error doesn't affect docker image pinning",
			filename:         "./testdata/Dockerfile-no-curl-sh-with-parser-error",
			processingErrors: 1, // everything is pinned, but parser error still throws warning
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.unpinned != unpinned {
				t.Errorf("expected %v unpinned. Got %v", tt.unpinned, unpinned)
			}

			if tt.processingErrors != len(r.ProcessingErrors) {
				t.Errorf("expected %v processing errors. Got %v", tt.processingErrors, len(r.ProcessingErrors))
			}
		})
	}
}

func TestShellScriptDownload(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name             string
		filename         string
		unpinned         int
		processingErrors int
		err              error
	}{
		{
			name:     "sh script",
			filename: "../testdata/script-sh",
			unpinned: 7,
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
			unpinned: 11,
		},
		{
			name:     "sh script 2",
			filename: "../testdata/script.sh",
			unpinned: 7,
		},
		{
			name:     "pkg managers",
			filename: "./testdata/script-pkg-managers",
			unpinned: 56,
		},
		{
			name:             "invalid shell script",
			filename:         "./testdata/script-invalid.sh",
			processingErrors: 1, // `curl bla | bash` not detected due to invalid script
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.unpinned != unpinned {
				t.Errorf("expected %v unpinned. Got %v", tt.unpinned, len(r.Dependencies))
			}

			if tt.processingErrors != len(r.ProcessingErrors) {
				t.Errorf("expected %v processing errors. Got %v", tt.processingErrors, len(r.ProcessingErrors))
			}
		})
	}
}

func TestShellScriptDownloadPinned(t *testing.T) {
	t.Parallel()
	//nolint:govet
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.warns != unpinned {
				t.Errorf("expected %v. Got %v", tt.warns, unpinned)
			}
		})
	}
}

func TestGitHubWorkflowRunDownload(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name             string
		filename         string
		unpinned         int
		processingErrors int
		err              error
	}{
		{
			name:     "workflow curl default",
			filename: "./testdata/.github/workflows/github-workflow-curl-default.yaml",
			unpinned: 1,
		},
		{
			name:     "workflow curl no default",
			filename: "./testdata/.github/workflows/github-workflow-curl-no-default.yaml",
			unpinned: 1,
		},
		{
			name:     "wget across steps",
			filename: "./testdata/.github/workflows/github-workflow-wget-across-steps.yaml",
			unpinned: 2,
		},
		{
			name:             "Can't identify OS but doesn't crash",
			filename:         "./testdata/.github/workflows/github-workflow-unknown-os.yaml",
			processingErrors: 1, // job with unknown OS has a skipped step
			unpinned:         1, // only found in 1 in job with known OS
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

			unpinned := countUnpinned(r.Dependencies)

			if tt.unpinned != unpinned {
				t.Errorf("expected %v unpinned. Got %v", tt.unpinned, unpinned)
			}

			if tt.processingErrors != len(r.ProcessingErrors) {
				t.Errorf("expected %v processing errors. Got %v", tt.processingErrors, len(r.ProcessingErrors))
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

func countUnpinned(r []checker.Dependency) int {
	var unpinned int

	for _, dependency := range r {
		if *dependency.Pinned == false {
			unpinned += 1
		}
	}

	return unpinned
}

func stringAsPointer(s string) *string {
	return &s
}

func boolAsPointer(b bool) *bool {
	return &b
}

// TestCollectDockerfilePinning tests the collectDockerfilePinning function.
func TestCollectDockerfilePinning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		filename            string
		outcomeDependencies []checker.Dependency
		expectError         bool
	}{
		{
			name:        "Workflow with error",
			filename:    "./testdata/.github/workflows/github-workflow-download-lines.yaml",
			expectError: true,
		},
		{
			name:        "Pinned dockerfile",
			filename:    "./testdata/Dockerfile-pinned",
			expectError: false,
			outcomeDependencies: []checker.Dependency{
				{
					Name:     stringAsPointer("python"),
					PinnedAt: stringAsPointer("3.7@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2"),
					Location: &checker.File{
						Path:      "./testdata/Dockerfile-pinned",
						Snippet:   "FROM python:3.7@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2",
						Offset:    16,
						EndOffset: 16,
						Type:      1,
					},
					Pinned: boolAsPointer(true),
					Type:   "containerImage",
				},
			},
		},
		{
			name:        "Non-pinned dockerfile",
			filename:    "./testdata/Dockerfile-not-pinned",
			expectError: false,
			outcomeDependencies: []checker.Dependency{
				{
					Name:     stringAsPointer("python"),
					PinnedAt: stringAsPointer("3.7"),
					Location: &checker.File{
						Path:      "./testdata/Dockerfile-not-pinned",
						Snippet:   "FROM python:3.7",
						Offset:    17,
						EndOffset: 17,
						FileSize:  0,
						Type:      1,
					},
					Pinned: boolAsPointer(false),
					Type:   "containerImage",
					Remediation: &finding.Remediation{
						Text: "pin your Docker image by updating python:3.7 to python:3.7" +
							"@sha256:eedf63967cdb57d8214db38ce21f105003ed4e4d0358f02bedc057341bcf92a0",
						Markdown: "pin your Docker image by updating python:3.7 to python:3.7" +
							"@sha256:eedf63967cdb57d8214db38ce21f105003ed4e4d0358f02bedc057341bcf92a0",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{tt.filename}, nil).AnyTimes()
			mockRepoClient.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()
			mockRepoClient.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open(file)
			})

			req := checker.CheckRequest{
				RepoClient: mockRepoClient,
			}
			var r checker.PinningDependenciesData
			err := collectDockerfilePinning(&req, &r)
			if err != nil {
				if !tt.expectError {
					t.Error(err.Error())
				}
			}
			for i := range tt.outcomeDependencies {
				outcomeDependency := &tt.outcomeDependencies[i]
				depend := &r.Dependencies[i]
				if diff := cmp.Diff(outcomeDependency, depend); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// TestCollectGitHubActionsWorkflowPinning tests the collectGitHubActionsWorkflowPinning function.
func TestCollectGitHubActionsWorkflowPinning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		filename            string
		outcomeDependencies []checker.Dependency
		expectError         bool
	}{
		{
			name:        "Pinned dockerfile",
			filename:    "Dockerfile-empty",
			expectError: true,
		},
		{
			name:        "Pinned workflow",
			filename:    ".github/workflows/workflow-pinned.yaml",
			expectError: false,
			outcomeDependencies: []checker.Dependency{
				{
					Name:     stringAsPointer("actions/checkout"),
					PinnedAt: stringAsPointer("daadedc81d5f9d3c06d2c92f49202a3cc2b919ba"),
					Location: &checker.File{
						Path:      ".github/workflows/workflow-pinned.yaml",
						Snippet:   "actions/checkout@daadedc81d5f9d3c06d2c92f49202a3cc2b919ba",
						Offset:    31,
						EndOffset: 31,
						Type:      1,
					},
					Pinned:      boolAsPointer(true),
					Type:        "GitHubAction",
					Remediation: nil,
				},
			},
		},
		{
			name:        "Non-pinned workflow",
			filename:    ".github/workflows/workflow-not-pinned.yaml",
			expectError: false,
			outcomeDependencies: []checker.Dependency{
				{
					Name:     stringAsPointer("actions/checkout"),
					PinnedAt: stringAsPointer("daadedc81d5f9d3c06d2c92f49202a3cc2b919ba"),
					Location: &checker.File{
						Path:      ".github/workflows/workflow-not-pinned.yaml",
						Snippet:   "actions/checkout@daadedc81d5f9d3c06d2c92f49202a3cc2b919ba",
						Offset:    31,
						EndOffset: 31,
						FileSize:  0,
						Type:      1,
					},
					Pinned:      boolAsPointer(true),
					Type:        "GitHubAction",
					Remediation: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{tt.filename}, nil).AnyTimes()
			mockRepoClient.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()
			mockRepoClient.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open(filepath.Join("testdata", file))
			})

			req := checker.CheckRequest{
				RepoClient: mockRepoClient,
			}
			var r checker.PinningDependenciesData
			err := collectGitHubActionsWorkflowPinning(&req, &r)
			if err != nil {
				if !tt.expectError {
					t.Error(err.Error())
				}
			}
			t.Log(r.Dependencies)
			for i := range tt.outcomeDependencies {
				outcomeDependency := &tt.outcomeDependencies[i]
				depend := &r.Dependencies[i]
				if diff := cmp.Diff(outcomeDependency, depend); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
