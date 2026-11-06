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
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyNpmLockFileV3TransitiveMissingIntegrity(t *testing.T) {
	t.Parallel()

	integrity := "sha512-" + strings.Repeat("A", 86) + "=="

	content := []byte(`{
		"name": "demo",
		"lockfileVersion": 3,
		"packages": {
			"": {
				"name": "demo",
				"version": "1.0.0"
			},
			"node_modules/foo": {
				"version": "1.0.0",
				"resolved": "https://registry.npmjs.org/foo/-/foo-1.0.0.tgz",
				"integrity": "` + integrity + `"
			},
			"node_modules/foo/node_modules/bar": {
				"version": "2.0.0",
				"resolved": "https://registry.npmjs.org/bar/-/bar-2.0.0.tgz"
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatalf("verifyNpmLockFile() error = %v", err)
	}

	if got.Applicable != 2 {
		t.Errorf("Applicable = %d, want 2", got.Applicable)
	}

	if len(got.Invalid) != 1 {
		t.Fatalf("len(Invalid) = %d, want 1", len(got.Invalid))
	}

	if got.Invalid[0] != "node_modules/foo/node_modules/bar" {
		t.Errorf(
			"Invalid[0] = %q, want %q",
			got.Invalid[0],
			"node_modules/foo/node_modules/bar",
		)
	}
}

func TestVerifyNpmLockFileV2(t *testing.T) {
	t.Parallel()
	content := []byte(`{
		"lockfileVersion": 2,
		"packages": {
			"": {
				"name": "fixture"
			},
			"node_modules/foo": {
				"version": "1.0.0",
				"integrity": "sha512-MJTUg1kjuLeQCJ+ccE4Vpa6kKVXkPYJ2mOCQyUuKLcLQsdrMCpBPUi8qVE6+YuaJkozeA9NusTAw3hLr8Xe5EQ=="
			},
			"node_modules/bar": {
				"version": "2.0.0"
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatalf("verifyNpmLockFile() error = %v", err)
	}

	if got.Applicable != 2 {
		t.Errorf("Applicable = %d, want 2", got.Applicable)
	}

	if len(got.Invalid) != 1 || got.Invalid[0] != "node_modules/bar" {
		t.Errorf("Invalid = %v, want [node_modules/bar]", got.Invalid)
	}
}

func TestVerifyNpmLockFileV1NestedDependency(t *testing.T) {
	t.Parallel()
	content := []byte(`{
		"lockfileVersion": 1,
		"dependencies": {
			"foo": {
				"version": "1.0.0",
				"integrity": "sha1-8S4PPF13sLHN2RRpQuTpbB5N1SU=",
				"dependencies": {
					"bar": {
						"version": "2.0.0"
					}
				}
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatalf("verifyNpmLockFile() error = %v", err)
	}

	if got.Applicable != 2 {
		t.Errorf("Applicable = %d, want 2", got.Applicable)
	}

	want := "foo/node_modules/bar"
	if len(got.Invalid) != 1 || got.Invalid[0] != want {
		t.Errorf("Invalid = %v, want [%s]", got.Invalid, want)
	}

	if len(got.Packages) != 2 {
		t.Fatalf("len(Packages) = %d, want 2", len(got.Packages))
	}
	if got.Packages[0].Path != "foo" || !got.Packages[0].Pinned {
		t.Errorf("Packages[0] = %+v, want pinned foo", got.Packages[0])
	}
	if got.Packages[1].Path != "foo/node_modules/bar" || got.Packages[1].Pinned {
		t.Errorf("Packages[1] = %+v, want unpinned nested bar", got.Packages[1])
	}
}

func TestVerifyNpmLockFileRejectsInvalidIntegrity(t *testing.T) {
	t.Parallel()
	content := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {
				"name": "fixture"
			},
			"node_modules/foo": {
				"version": "1.0.0",
				"integrity": "not-an-sri"
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatalf("verifyNpmLockFile() error = %v", err)
	}

	if got.Applicable != 1 {
		t.Errorf("Applicable = %d, want 1", got.Applicable)
	}

	if len(got.Invalid) != 1 || got.Invalid[0] != "node_modules/foo" {
		t.Errorf("Invalid = %v, want [node_modules/foo]", got.Invalid)
	}
}

func TestVerifyNpmLockFileIgnoresRootLinkBundle(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {
				"name": "fixture"
			},
			"node_modules/link": {
				"link": true,
				"resolved": "packages/link"
			},
			"packages/link": {},
			"node_modules/bundled": {
				"inBundle": true
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatalf("verifyNpmLockFile() error = %v", err)
	}

	if got.Applicable != 0 {
		t.Errorf("Applicable = %d, want 0", got.Applicable)
	}

	if len(got.Invalid) != 0 {
		t.Errorf("Invalid = %v, want []", got.Invalid)
	}
}

func TestVerifyNpmLockFileMalformedJSON(t *testing.T) {
	t.Parallel()

	_, err := verifyNpmLockFile([]byte(`{"lockfileVersion": 3`))
	if err == nil {
		t.Fatal("verifyNpmLockFile() error = nil, want error")
	}
}

func TestVerifyNpmLockFileUnsupportedVersion(t *testing.T) {
	t.Parallel()

	_, err := verifyNpmLockFile([]byte(`{
		"lockfileVersion": 99,
		"packages": {}
	}`))
	if err == nil {
		t.Fatal("verifyNpmLockFile() error = nil, want error")
	}
}

func TestValidNpmIntegrity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		integrity string
		want      bool
	}{
		{
			name:      "valid sha1",
			integrity: "sha1-" + strings.Repeat("A", 27) + "=",
			want:      true,
		},
		{
			name:      "valid sha512",
			integrity: "sha512-" + strings.Repeat("A", 86) + "==",
			want:      true,
		},
		{
			name:      "empty",
			integrity: "",
			want:      false,
		},
		{
			name:      "invalid base64",
			integrity: "sha512-not-base64!!!",
			want:      false,
		},
		{
			name:      "wrong sha1 length",
			integrity: "sha1-AAAA",
			want:      false,
		},
		{
			name:      "unsupported algorithm",
			integrity: "md5-" + strings.Repeat("A", 22) + "==",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := validNpmIntegrity(tt.integrity)
			if got != tt.want {
				t.Errorf(
					"validNpmIntegrity(%q) = %v, want %v",
					tt.integrity,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestVerifyNpmLockFileV3GitCommitPinned(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {"name": "fixture"},
			"node_modules/git-package": {
				"version": "1.0.0",
				"resolved": "git+https://github.com/example/git-package.git#0123456789abcdef0123456789abcdef01234567"
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatal(err)
	}
	wantPath := "node_modules/git-package"
	if got.Applicable != 1 || len(got.Invalid) != 0 ||
		len(got.Packages) != 1 ||
		got.Packages[0].Path != wantPath ||
		!got.Packages[0].Pinned {
		t.Errorf("Git commit should be pinned; got %+v", got)
	}
}

func TestVerifyNpmLockFileV1GitCommitPinned(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 1,
		"dependencies": {
			"git-package": {
				"version": "git+https://github.com/example/git-package.git#0123456789abcdef0123456789abcdef01234567"
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatal(err)
	}
	wantPath := "git-package"
	if got.Applicable != 1 || len(got.Invalid) != 0 ||
		len(got.Packages) != 1 ||
		got.Packages[0].Path != wantPath ||
		!got.Packages[0].Pinned {
		t.Errorf("Git commit should be pinned; got %+v", got)
	}
}

func TestVerifyNpmLockFileV3WorkspaceLink(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {"name": "fixture"},
			"node_modules/lib": {
				"resolved": "packages/lib",
				"link": true
			},
			"packages/lib": {
				"version": "1.0.0"
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatal(err)
	}
	if got.Applicable != 0 || len(got.Invalid) != 0 || len(got.Packages) != 0 {
		t.Errorf("workspace link should be outside SRI scope; got %+v", got)
	}
}

func TestVerifyNpmLockFileV1LocalDirectoryOutsideSRIScope(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 1,
		"dependencies": {
			"local-package": {
				"version": "file:libs/local-package"
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatal(err)
	}
	if got.Applicable != 0 || len(got.Invalid) != 0 || len(got.Packages) != 0 {
		t.Errorf("local directory should be outside SRI scope; got %+v", got)
	}
}

func TestVerifyNpmLockFileV1LocalTarballMissingIntegrity(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 1,
		"dependencies": {
			"local-pkg": {
				"version": "file:local-pkg-1.0.0.tgz"
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatal(err)
	}
	if got.Applicable != 1 || len(got.Invalid) != 1 ||
		got.Invalid[0] != "local-pkg" {
		t.Errorf("local tarball without integrity should fail; got %+v", got)
	}
}

func TestVerifyNpmLockFileV3GitBranchUnpinned(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {"name": "fixture"},
			"node_modules/git-package": {
				"resolved": "git+https://github.com/example/git-package.git#main",
				"integrity": "sha512-MJTUg1kjuLeQCJ+ccE4Vpa6kKVXkPYJ2mOCQyUuKLcLQsdrMCpBPUi8qVE6+YuaJkozeA9NusTAw3hLr8Xe5EQ=="
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatal(err)
	}
	if got.Applicable != 1 || len(got.Invalid) != 1 ||
		got.Invalid[0] != "node_modules/git-package" ||
		len(got.Packages) != 1 || got.Packages[0].Pinned {
		t.Errorf("Git branch should be unpinned; got %+v", got)
	}
}

func TestMissingNpmLockDependencies(t *testing.T) {
	t.Parallel()

	manifest := []byte(`{
		"dependencies": {
			"foo": "^1.0.0",
			"missing": "^2.0.0"
		},
		"devDependencies": {
			"@scope/tool": "^1.0.0",
			"missing": "^2.0.0"
		},
		"optionalDependencies": {
			"optional-missing": "^1.0.0"
		}
	}`)

	tests := []struct {
		name string
		lock string
	}{
		{
			name: "v1",
			lock: `{
				"lockfileVersion": 1,
				"dependencies": {
					"foo": {},
					"@scope/tool": {}
				}
			}`,
		},
		{
			name: "v2",
			lock: `{
				"lockfileVersion": 2,
				"packages": {
					"": {},
					"node_modules/foo": {"link": true},
					"node_modules/@scope/tool": {"inBundle": true}
				}
			}`,
		},
		{
			name: "v3",
			lock: `{
				"lockfileVersion": 3,
				"packages": {
					"": {},
					"node_modules/foo": {"link": true},
					"node_modules/@scope/tool": {"inBundle": true}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := missingNpmLockDependencies(manifest, []byte(tt.lock))
			if err != nil {
				t.Fatal(err)
			}

			if len(got) != 2 ||
				got[0] != "missing" ||
				got[1] != "optional-missing" {
				t.Errorf("missing = %v, want [missing optional-missing]", got)
			}
		})
	}
}

func TestMissingNpmProjectDependenciesWorkspace(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	manifests := map[string][]byte{
		projectDir: []byte(`{
			"dependencies": {"lib": "file:packages/lib"}
		}`),
		filepath.Join(projectDir, "packages", "lib"): []byte(`{
			"dependencies": {
				"local": "^1.0.0",
				"hoisted": "^1.0.0",
				"@scope/missing": "^1.0.0"
			}
		}`),
	}

	lock := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {},
			"node_modules/lib": {
				"link": true,
				"resolved": "packages/lib"
			},
			"packages/lib": {},
			"packages/lib/node_modules/local": {},
			"node_modules/hoisted": {}
		}
	}`)

	got, err := missingNpmProjectDependencies(lock, projectDir, manifests)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatalf("missing = %+v, want exactly one dependency", got)
	}
	if got[0].Path != "packages/lib/node_modules/@scope/missing" ||
		got[0].Pinned {
		t.Errorf("unexpected missing dependency: %+v", got[0])
	}
}

func TestMissingNpmProjectDependenciesAbsentWorkspace(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	manifests := map[string][]byte{
		projectDir: []byte(`{
			"workspaces": ["./packages/*"]
		}`),
		filepath.Join(projectDir, "packages", "lib"): []byte(`{
			"dependencies": {"foo": "^1.0.0"}
		}`),
		filepath.Join(projectDir, "examples", "demo"): []byte(`{
			"dependencies": {"unrelated": "^1.0.0"}
		}`),
	}

	lock := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {},
			"node_modules/foo": {}
		}
	}`)

	got, err := missingNpmProjectDependencies(lock, projectDir, manifests)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatalf("missing = %+v, want only the absent workspace", got)
	}
	if got[0].Path != "packages/lib" || got[0].Pinned {
		t.Errorf("unexpected missing workspace: %+v", got[0])
	}
}

func TestVerifyNpmLockFileLinkMissingTarget(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {},
			"node_modules/foo": {
				"link": true,
				"resolved": "packages/foo",
				"integrity": "sha512-` + strings.Repeat("A", 86) + `=="
			}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatal(err)
	}

	if got.Applicable != 1 ||
		len(got.Invalid) != 1 ||
		got.Invalid[0] != "node_modules/foo" ||
		len(got.Packages) != 1 ||
		got.Packages[0].Path != "node_modules/foo" ||
		got.Packages[0].Pinned {
		t.Fatalf("link without target should be unpinned; got %+v", got)
	}
}

func TestVerifyNpmLockFileLinkDoesNotHideInstalledPackage(t *testing.T) {
	t.Parallel()

	content := []byte(`{
		"lockfileVersion": 3,
		"packages": {
			"": {},
			"node_modules/alias": {
				"link": true,
				"resolved": "node_modules/foo"
			},
			"node_modules/foo": {}
		}
	}`)

	got, err := verifyNpmLockFile(content)
	if err != nil {
		t.Fatal(err)
	}

	if got.Applicable != 1 ||
		len(got.Invalid) != 1 ||
		got.Invalid[0] != "node_modules/foo" ||
		len(got.Packages) != 1 ||
		got.Packages[0].Path != "node_modules/foo" ||
		got.Packages[0].Pinned {
		t.Fatalf("link must not hide unpinned foo; got %+v", got)
	}
}
