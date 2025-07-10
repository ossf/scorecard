// Copyright 2025 OpenSSF Scorecard Authors
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
	stdOs "os"
	"path/filepath"
	"testing"

	"deps.dev/util/semver"
)

type testCase struct {
	lockFile         string
	renameLockFileTo string
	expected         []*Dependency
}

func TestScanLockFile(t *testing.T) {
	t.Parallel()
	testCases := []*testCase{
		{
			lockFile:         "reduced-scorecard.go.mod",
			renameLockFileTo: "go.mod",
			expected: []*Dependency{
				{
					Name:    "cloud.google.com/go/bigquery",
					Version: "1.69.0",
				},
				{
					Name:    "cloud.google.com/go/monitoring",
					Version: "1.24.2",
				},
				{
					Name:    "cloud.google.com/go/pubsub",
					Version: "1.49.0",
				},
				{
					Name:    "cloud.google.com/go/trace",
					Version: "1.11.6",
				},
				{
					Name:    "contrib.go.opencensus.io/exporter/stackdriver",
					Version: "0.13.14",
				},
				{
					Name:    "github.com/bombsimon/logrusr/v2",
					Version: "2.0.1",
				},
				{
					Name:    "stdlib",
					Version: "1.23.0",
				},
			},
		},
		{
			lockFile:         "reduced-jackson-core.pom.xml",
			renameLockFileTo: "pom.xml",
			expected: []*Dependency{
				{
					Name:    "ch.randelshofer:fastdoubleparser",
					Version: "2.0.1",
				},
				{
					Name: "org.junit.jupiter:junit-jupiter",
				},
				{
					Name: "org.junit.jupiter:junit-jupiter-api",
				},
				{
					Name:    "org.assertj:assertj-core",
					Version: "",
				},
			},
		},
		{
			lockFile:         "reduced-fuzz-introspector.requirements.txt",
			renameLockFileTo: "requirements.txt",
			expected: []*Dependency{
				{
					Name:       "beautifulsoup4",
					Version:    "4.10.0",
					Comparator: "==",
				},
				{
					Name:       "cxxfilt",
					Version:    "0.3.0",
					Comparator: "==",
				},
				{
					Name:       "lxml",
					Version:    "5.3.0",
					Comparator: "==",
				},
				{
					Name:       "matplotlib",
					Version:    "3.10.0",
					Comparator: "==",
				},
				{
					Name:       "numpy",
					Version:    "2.1.0",
					Comparator: "==",
				},
				{
					Name:       "PyYAML",
					Version:    "6.0.2",
					Comparator: "==",
				},
				{
					Name:    "flake8",
					Version: "",
				},
			},
		},
		{
			lockFile:         "pytorch.requirements.txt",
			renameLockFileTo: "requirements.txt",
			expected: []*Dependency{
				{
					Name: "setuptools",
				},
				{
					Name:       "cmake",
					Version:    "3.27",
					Comparator: ">=",
				},
				{
					Name: "ninja",
				},
				{
					Name: "numpy",
				},
				{
					Name: "packaging",
				},
				{
					Name: "pyyaml",
				},
				{
					Name: "requests",
				},
				{
					Name: "six",
				},
				{
					Name:       "typing-extensions",
					Version:    "4.10.0",
					Comparator: ">=",
				},
				{
					Name: "build",
				},
				{
					Name:       "expecttest",
					Version:    "0.3.0",
					Comparator: ">=",
				},
				{
					Name: "filelock",
				},
				{
					Name: "fsspec",
				},
				{
					Name: "hypothesis",
				},
				{
					Name: "jinja2",
				},
				{
					Name: "lintrunner",
				},
				{
					Name: "networkx",
				},
				{
					Name:       "optree",
					Version:    "0.13.0",
					Comparator: ">=",
				},
				{
					Name: "psutil",
				},
				{
					Name:       "sympy",
					Version:    "1.13.3",
					Comparator: ">=",
				},
			},
		},
		{
			lockFile:         "reduced-colors.package-lock.json",
			renameLockFileTo: "package-lock.json",
			expected: []*Dependency{
				{
					Name:    "@babel/code-frame",
					Version: "7.0.0",
				},
				{
					Name:    "@babel/highlight",
					Version: "7.0.0",
				},
				{
					Name:    "acorn",
					Version: "6.0.4",
				},
				{
					Name:    "acorn-jsx",
					Version: "5.0.1",
				},
				{
					Name:    "ajv",
					Version: "6.6.1",
				},
				{
					Name:    "ansi-escapes",
					Version: "3.1.0",
				},
				{
					Name:    "ansi-regex",
					Version: "3.0.0",
				},
				{
					Name:    "ansi-styles",
					Version: "3.2.1",
				},
				{
					Name:    "argparse",
					Version: "1.0.10",
				},
			},
		},
		{
			lockFile:         "reduced-fastify.package.json",
			renameLockFileTo: "package.json",
			expected: []*Dependency{
				{
					Name:    "@fastify/ajv-compiler",
					Version: "^4.0.0",
				},
				{
					Name:    "@fastify/error",
					Version: "^4.0.0",
				},
				{
					Name:    "@fastify/fast-json-stringify-compiler",
					Version: "^5.0.0",
				},
				{
					Name:    "@fastify/proxy-addr",
					Version: "^5.0.0",
				},
				{
					Name:    "abstract-logging",
					Version: "^2.0.1",
				},
				{
					Name:    "avvio",
					Version: "^9.0.0",
				},
				{
					Name:    "fast-json-stringify",
					Version: "^6.0.0",
				},
				{
					Name:    "find-my-way",
					Version: "^9.0.0",
				},
				{
					Name:    "light-my-request",
					Version: "^6.0.0",
				},
				{
					Name:    "pino",
					Version: "^9.0.0",
				},
				{
					Name:    "process-warning",
					Version: "^5.0.0",
				},
				{
					Name:    "rfdc",
					Version: "^1.3.1",
				},
				{
					Name:    "secure-json-parse",
					Version: "^4.0.0",
				},
				{
					Name:    "semver",
					Version: "^7.6.0",
				},
				{
					Name:    "toad-cache",
					Version: "^3.7.0",
				},
			},
		},
		{
			lockFile:         "package-lock.json",
			renameLockFileTo: "package-lock.json",
			expected: []*Dependency{
				{
					Name:    "@angular/animations",
					Version: "4.2.6",
				},
				{
					Name:    "@angular/common",
					Version: "4.2.6",
				},
				{
					Name:    "@angular/compiler",
					Version: "4.2.6",
				},
				{
					Name:    "@angular/core",
					Version: "4.2.6",
				},
				{
					Name:    "@angular/forms",
					Version: "4.2.6",
				},
				{
					Name:    "@angular/http",
					Version: "4.2.6",
				},
				{
					Name:    "@angular/platform-browser",
					Version: "4.2.7",
				},
			},
		},
	}
	for _, tt := range testCases {
		fileContents, err := stdOs.ReadFile(filepath.Join("testdata", tt.lockFile))
		if err != nil {
			t.Fatalf("could not read file %s, err: %v", tt.lockFile, err)
		}
		// create a temp dir and make a copy of the lock file.
		// we need the temp dir in the current dir because
		// of how osv-scalibr works.
		//nolint:usetesting
		tmpDir, err := stdOs.MkdirTemp(".", "lockfile-test")
		if err != nil {
			t.Fatalf("could not create a temporary directory: %v", err)
		}
		defer func() {
			err := stdOs.RemoveAll(tmpDir)
			if err != nil {
				t.Errorf("could not delete temp test dir: %v", err)
			}
		}()
		tmpLockFile := filepath.Join(tmpDir, tt.renameLockFileTo)
		f, err := stdOs.Create(tmpLockFile)
		if err != nil {
			t.Fatalf("could not create temporary lockfile, err: %v", err)
		}
		defer f.Close()
		_, err = f.Write(fileContents)
		if err != nil {
			t.Fatalf("could not write file: %v", err)
		}

		deps, err := ScanLockFile(tmpLockFile)
		if err != nil {
			t.Fatalf("got err: %v", err)
		}
		// check that we found all the depdencies we expected
		checkExpectedDeps(t, tt.expected, deps)

		// check that we only got the depdencies that we expected.
		checkGotOnlyExpectedDeps(t, tt.lockFile, tt.expected, deps)
	}
}

func checkGotOnlyExpectedDeps(t *testing.T, lockFile string, expected, got []*Dependency) {
	t.Helper()
	if len(got) > len(expected) {
		t.Logf("%s: got more dependencies than expected", lockFile)
		t.Logf("expected deps: %d, len(deps): %d",
			len(expected), len(got))
	} else {
		return
	}
	for _, dep := range got {
		if !weExpectThisDependency(dep, expected) {
			// ScanLockFile returned a dependency that we did not expect
			t.Errorf("%s: got unexpected dependency '%s' version '%s' comparator '%s'",
				lockFile, dep.Name, dep.Version, dep.Comparator)
		}
	}
}

func checkExpectedDeps(t *testing.T, expected, got []*Dependency) {
	t.Helper()
	for i := range expected {
		// check if ScanLockFile found all expected dependencies
		p := getMatchingDep(got, expected[i].Name)
		// linter complains because of a potential nil-pointer dereference.
		//nolint:staticcheck
		if p == nil {
			t.Errorf("could not find dependency: %s", expected[i].Name)
		}
		// linter complains because of a potential nil-pointer dereference.
		//nolint:staticcheck
		if expected[i].Version != p.Version {
			// this is useful for debugging
			t.Logf("found %s which is expected, however, the expected version is %s and got %s",
				p.Name, expected[i].Version, p.Version)
			continue
		}

		// For python-requirements, we check the version comparator
		if expected[i].Comparator != p.Comparator {
			t.Errorf("found %s with the right version, but expected Comparator %s and got %s",
				expected[i].Name, expected[i].Comparator, p.Comparator)
		}
	}
}

func weExpectThisDependency(dep *Dependency, expectedDeps []*Dependency) bool {
	expect := false
	for i := range len(expectedDeps) {
		if dep.Name == expectedDeps[i].Name &&
			dep.Version == expectedDeps[i].Version &&
			dep.Comparator == expectedDeps[i].Comparator {
			expect = true
			break
		}
	}
	return expect
}

func getMatchingDep(deps []*Dependency, depName string) *Dependency {
	for _, dep := range deps {
		if depName != dep.Name {
			continue
		}
		return dep
	}
	return nil
}

func TestCompare(t *testing.T) {
	t.Parallel()
	type compareTest struct {
		dep1     *Dependency
		dep2     *Dependency
		name     string
		expected bool
	}
	testCases := []*compareTest{
		{
			name: "maven: expect true because version is higher",
			dep1: &Dependency{
				Ecosystem: semver.Maven,
				Version:   "2.19.1",
			},
			dep2: &Dependency{
				Ecosystem: semver.Maven,
				Version:   "2.19.0",
			},
			expected: true,
		},
		{
			name: "maven: expect true because version is higher than rc",
			dep1: &Dependency{
				Ecosystem: semver.Maven,
				Version:   "2.19.0",
			},
			dep2: &Dependency{
				Ecosystem: semver.Maven,
				Version:   "2.19.0-rc2",
			},
			expected: true,
		},
		{
			name: "maven: expect true because rc version is higher",
			dep1: &Dependency{
				Ecosystem: semver.Maven,
				Version:   "2.19.0-rc2",
			},
			dep2: &Dependency{
				Ecosystem: semver.Maven,
				Version:   "2.18.4",
			},
			expected: true,
		},
		{
			name: "go: expect true because version is higher",
			dep1: &Dependency{
				Ecosystem: semver.Go,
				Version:   "v5.16.2",
			},
			dep2: &Dependency{
				Ecosystem: semver.Go,
				Version:   "v5.13.2",
			},
			expected: true,
		},
		{
			name: "npm: expect false because version is lower than range",
			dep1: &Dependency{
				Ecosystem: semver.NPM,
				Version:   "3.0.0",
			},
			dep2: &Dependency{
				Ecosystem: semver.NPM,
				Version:   "^4.0.0",
			},
			expected: false,
		},
		{
			name: "npm: expect false because version is higher than range",
			dep1: &Dependency{
				Ecosystem: semver.NPM,
				Version:   "5.0.0",
			},
			dep2: &Dependency{
				Ecosystem: semver.NPM,
				Version:   "^4.0.0",
			},
			expected: false,
		},
		{
			name: "npm: expect true because version is in range",
			dep1: &Dependency{
				Ecosystem: semver.NPM,
				Version:   "4.1.0",
			},
			dep2: &Dependency{
				Ecosystem: semver.NPM,
				Version:   "^4.0.0",
			},
			expected: true,
		},
		{
			name: "python: expect true because version is in range",
			dep1: &Dependency{
				Ecosystem: semver.PyPI,
				Version:   "70.1.1",
			},
			dep2: &Dependency{
				Ecosystem: semver.PyPI,
				Version:   ">=70.1.0,<80.0",
			},
			expected: true,
		},
		{
			name: "python: expect false because version is higher than range",
			dep1: &Dependency{
				Ecosystem: semver.PyPI,
				Version:   "90.1.1",
			},
			dep2: &Dependency{
				Ecosystem: semver.PyPI,
				Version:   ">=70.1.0,<80.0",
			},
			expected: false,
		},
	}
	for _, tt := range testCases {
		got := tt.dep1.Compare(tt.dep2)
		if got != tt.expected {
			t.Errorf("%s: %s returned %t but expected %t", tt.name, tt.dep1.Name, got, tt.expected)
		}
	}
}
