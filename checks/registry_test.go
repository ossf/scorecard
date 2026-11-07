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

package checks

import (
	"maps"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/checker"
)

// registeredCheck is what the registry records for one check: the
// function it runs, by name, and its supported request types.
type registeredCheck struct {
	Fn           string
	RequestTypes []checker.RequestType
}

var (
	commitBased        = []checker.RequestType{checker.CommitBased}
	fileBased          = []checker.RequestType{checker.FileBased}
	commitAndFileBased = []checker.RequestType{checker.CommitBased, checker.FileBased}
)

// wantRegistry is every registered check, experimental ones included.
// Nil request types are kept apart from empty ones.
var wantRegistry = map[string]registeredCheck{
	"Binary-Artifacts":       {Fn: "BinaryArtifacts", RequestTypes: commitAndFileBased},
	"Branch-Protection":      {Fn: "BranchProtection"},
	"CI-Tests":               {Fn: "CITests", RequestTypes: commitBased},
	"CII-Best-Practices":     {Fn: "CIIBestPractices"},
	"Code-Review":            {Fn: "CodeReview", RequestTypes: commitBased},
	"Contributors":           {Fn: "Contributors"},
	"Dangerous-Workflow":     {Fn: "DangerousWorkflow", RequestTypes: commitAndFileBased},
	"Dependency-Update-Tool": {Fn: "DependencyUpdateTool", RequestTypes: fileBased},
	"Fuzzing":                {Fn: "Fuzzing", RequestTypes: fileBased},
	"License":                {Fn: "License", RequestTypes: commitAndFileBased},
	"Maintained":             {Fn: "Maintained"},
	"Packaging":              {Fn: "Packaging", RequestTypes: fileBased},
	"Pinned-Dependencies":    {Fn: "PinningDependencies", RequestTypes: commitAndFileBased},
	"SAST":                   {Fn: "SAST", RequestTypes: fileBased},
	"SBOM":                   {Fn: "SBOM"},
	"Security-Policy":        {Fn: "SecurityPolicy", RequestTypes: commitAndFileBased},
	"Signed-Releases":        {Fn: "SignedReleases"},
	"Token-Permissions":      {Fn: "TokenPermissions", RequestTypes: commitAndFileBased},
	"Vulnerabilities":        {Fn: "Vulnerabilities", RequestTypes: commitAndFileBased},
	"Webhooks":               {Fn: "WebHooks"},
}

// experimentalChecks run only when SCORECARD_EXPERIMENTAL is set.
var experimentalChecks = []string{"SBOM", "Webhooks"}

// snapshot records each check's function by name, relative to this
// package, and its request types.
func snapshot(t *testing.T, m checker.CheckNameToFnMap) map[string]registeredCheck {
	t.Helper()
	const pkg = "github.com/ossf/scorecard/v5/checks."
	got := make(map[string]registeredCheck, len(m))
	for name, c := range m {
		fn := runtime.FuncForPC(reflect.ValueOf(c.Fn).Pointer()).Name()
		local, ok := strings.CutPrefix(fn, pkg)
		if !ok {
			t.Errorf("%s runs %q, not a function in package checks", name, fn)
		}
		got[name] = registeredCheck{Fn: local, RequestTypes: c.SupportedRequestTypes}
	}
	return got
}

func TestRegistry(t *testing.T) {
	t.Parallel()
	got := snapshot(t, GetAllWithExperimental())
	if diff := cmp.Diff(wantRegistry, got); diff != "" {
		t.Errorf("registry (-want +got):\n%s", diff)
	}
}

func TestGetAllExperimental(t *testing.T) {
	all := slices.Sorted(maps.Keys(wantRegistry))
	stable := slices.DeleteFunc(slices.Clone(all), func(name string) bool {
		return slices.Contains(experimentalChecks, name)
	})
	tests := []struct {
		name             string
		env              string
		want             []string
		wantExperimental []string
		set              bool
	}{
		{name: "unset", want: stable, wantExperimental: all},
		{name: "set", set: true, env: "1", want: all, wantExperimental: all},
		{name: "set to empty", set: true, env: "", want: all, wantExperimental: all},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SCORECARD_EXPERIMENTAL", tt.env)
			if !tt.set {
				unsetExperimental(t)
			}
			if diff := cmp.Diff(tt.want, slices.Sorted(maps.Keys(GetAll()))); diff != "" {
				t.Errorf("GetAll (-want +got):\n%s", diff)
			}
			got := slices.Sorted(maps.Keys(GetAllWithExperimental()))
			if diff := cmp.Diff(tt.wantExperimental, got); diff != "" {
				t.Errorf("GetAllWithExperimental (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetAllReturnsCopy(t *testing.T) {
	t.Parallel()
	for _, get := range []func() checker.CheckNameToFnMap{GetAll, GetAllWithExperimental} {
		m := get()
		clear(m)
		m["Extra"] = checker.Check{}
	}
	got := snapshot(t, GetAllWithExperimental())
	if diff := cmp.Diff(wantRegistry, got); diff != "" {
		t.Errorf("registry after callers changed returned maps (-want +got):\n%s", diff)
	}
}
