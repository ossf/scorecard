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

package fileparser

import (
	stdos "os"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/internal/ghworkflow"
)

func TestClassifyPublishAuth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		filename  string
		workflow  string
		auth      checker.PublishAuth
		ecosystem string
	}{
		{
			name:      "pypi token in existing fixture",
			filename:  "../testdata/.github/workflows/github-workflow-packaging-pypi.yaml",
			auth:      checker.PublishAuthCredential,
			ecosystem: ecosystemPyPI,
		},
		{
			name: "pypi oidc",
			workflow: `
on: push
permissions:
  id-token: write
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: pypa/gh-action-pypi-publish@release/v1
`,
			auth:      checker.PublishAuthOIDC,
			ecosystem: ecosystemPyPI,
		},
		{
			name: "pypi oidc permission is replaced by the job",
			workflow: `
on: push
permissions:
  id-token: write
jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: pypa/gh-action-pypi-publish@release/v1
`,
			auth:      checker.PublishAuthCredential,
			ecosystem: ecosystemPyPI,
		},
		{
			name: "pypi password wins over id-token",
			workflow: `
on: push
permissions:
  id-token: write
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: pypa/gh-action-pypi-publish@release/v1
        with:
          password: ${{ secrets.PYPI_API_TOKEN }}
`,
			auth:      checker.PublishAuthCredential,
			ecosystem: ecosystemPyPI,
		},
		{
			name: "pypi write-all counts as id-token write",
			workflow: `
on: push
jobs:
  publish:
    runs-on: ubuntu-latest
    permissions: write-all
    steps:
      - uses: pypa/gh-action-pypi-publish@release/v1
`,
			auth:      checker.PublishAuthOIDC,
			ecosystem: ecosystemPyPI,
		},
		{
			name:      "npm token in existing fixture",
			filename:  "../testdata/.github/workflows/github-workflow-packaging-npm.yaml",
			auth:      checker.PublishAuthCredential,
			ecosystem: ecosystemNPM,
		},
		{
			name: "npm oidc",
			workflow: `
on: push
permissions:
  id-token: write
  contents: read
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v4
        with:
          registry-url: https://registry.npmjs.org
      - run: npm publish
`,
			auth:      checker.PublishAuthOIDC,
			ecosystem: ecosystemNPM,
		},
		{
			name: "npm token on the publish step is not oidc",
			workflow: `
on: push
permissions:
  id-token: write
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v4
        with:
          registry-url: https://registry.npmjs.org
      - run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
`,
			auth:      checker.PublishAuthCredential,
			ecosystem: ecosystemNPM,
		},
		{
			name:      "gem push without oidc",
			filename:  "../testdata/.github/workflows/github-workflow-packaging-gem.yaml",
			auth:      checker.PublishAuthCredential,
			ecosystem: ecosystemRubyGems,
		},
		{
			name: "rubygems trusted publishing",
			workflow: `
on: push
permissions:
  id-token: write
  contents: write
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: rubygems/release-gem@v1
`,
			auth:      checker.PublishAuthOIDC,
			ecosystem: ecosystemRubyGems,
		},
		{
			name:      "cargo publish without oidc",
			filename:  "../testdata/.github/workflows/github-workflow-packaging-cargo.yaml",
			auth:      checker.PublishAuthCredential,
			ecosystem: ecosystemCrates,
		},
		{
			name: "crates trusted publishing",
			workflow: `
on: push
permissions:
  id-token: write
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - run: cargo publish
`,
			auth:      checker.PublishAuthOIDC,
			ecosystem: ecosystemCrates,
		},
		{
			name: "pub.dev trusted publishing",
			workflow: `
on: push
permissions:
  id-token: write
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - run: dart pub publish --force
`,
			auth:      checker.PublishAuthOIDC,
			ecosystem: ecosystemPub,
		},
		{
			name: "docker push is not classified",
			workflow: `
on: push
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - run: docker push ghcr.io/example/app:latest
`,
			auth:      "",
			ecosystem: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content := []byte(tt.workflow)
			if tt.filename != "" {
				var err error
				content, err = stdos.ReadFile(tt.filename)
				if err != nil {
					t.Fatalf("reading %s: %v", tt.filename, err)
				}
			}
			workflow, errs := ghworkflow.Parse(content)
			if workflow == nil {
				t.Fatalf("parse failed: %v", errs)
			}
			auth, ecosystem := ClassifyPublishAuth(workflow)
			if auth != tt.auth || ecosystem != tt.ecosystem {
				t.Fatalf("ClassifyPublishAuth() = (%q, %q), want (%q, %q)", auth, ecosystem, tt.auth, tt.ecosystem)
			}
		})
	}
}

func TestIsPackagingWorkflowTrustedPublishers(t *testing.T) {
	t.Parallel()
	content := []byte(`
on: push
jobs:
  rubygems:
    runs-on: ubuntu-latest
    steps:
      - uses: rubygems/release-gem@v1
  pub:
    runs-on: ubuntu-latest
    steps:
      - run: flutter pub publish --force
`)
	workflow, errs := ghworkflow.Parse(content)
	if workflow == nil {
		t.Fatalf("parse failed: %v", errs)
	}
	if _, ok := IsPackagingWorkflow(workflow, ".github/workflows/release.yml"); !ok {
		t.Fatal("IsPackagingWorkflow() = false, want true")
	}
}
