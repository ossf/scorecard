// Copyright 2022 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package clients

import (
	"reflect"
	"slices"
	"testing"

	transitiverequirements "github.com/google/osv-scalibr/enricher/transitivedependency/requirements"
	"github.com/google/osv-scanner/v2/pkg/models"
	"github.com/ossf/osv-schema/bindings/go/osvschema"
)

func TestRemoveDuplicate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		keyExtract func(string) string
		list       []string
		want       []string
	}{
		{
			name: "Basic list with dup items",
			list: []string{"A", "B", "C", "B"},
			want: []string{"A", "B", "C"},
			keyExtract: func(in string) string {
				return in
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := removeDuplicate(tt.list, tt.keyExtract)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmptyProject(t *testing.T) {
	t.Parallel()
	var client osvClient
	var commit string
	emptyDir := t.TempDir()
	_, err := client.ListUnfixedVulnerabilities(t.Context(), commit, emptyDir)
	if err != nil {
		t.Fatalf("empty directory shouldn't throw an error: %v", err)
	}
}

func TestPythonTransitivePluginName(t *testing.T) {
	t.Parallel()

	actions := (osvClient{}).scannerActions(nil, nil)
	disabled := actions.PluginsDisabled
	if !slices.Contains(disabled, transitiverequirements.Name) {
		t.Fatal("Python transitive requirements plugin must be disabled by its registered name")
	}
	if actions.TransitiveScanning.Disabled {
		t.Fatal("disabling Python transitive resolution must not disable other ecosystems")
	}
}

func TestLocalClientDisablesTransitiveScanning(t *testing.T) {
	t.Parallel()

	actions := (osvClient{local: true}).scannerActions(nil, nil)
	if !actions.TransitiveScanning.Disabled {
		t.Fatal("local client must disable transitive scanning to stay offline")
	}
}

func TestCollectVulnerabilities(t *testing.T) {
	t.Parallel()
	result := func(ecosystem, name, version, id string) models.VulnerabilityFlattened {
		return models.VulnerabilityFlattened{
			Package:       models.PackageInfo{Ecosystem: ecosystem, Name: name, Version: version},
			Vulnerability: &osvschema.Vulnerability{Id: id},
		}
	}
	vulns := []models.VulnerabilityFlattened{
		result("Go", "stdlib", "1.22.0", "GO-STDLIB"),
		result("Maven", "org.apache.tomcat:tomcat-juli", "@MAVEN.DEPLOY.VERSION@", "GHSA-PLACEHOLDER"),
		result("Maven", "org.apache.logging.log4j:log4j-core", "2.14.1", "GHSA-REAL"),
		result("npm", "tar", "4.4.8", "GHSA-NPM"),
		result("npm", "tar", "4.4.13", "GHSA-NPM"),
	}
	var got []string
	for _, v := range collectVulnerabilities(vulns) {
		got = append(got, v.ID)
	}
	want := []string{"GHSA-REAL", "GHSA-NPM"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestIsPlaceholderVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		version string
		want    bool
	}{
		{version: "@MAVEN.DEPLOY.VERSION@", want: true},
		{version: "${project.version}", want: true},
		{version: "1.0-${revision}", want: true},
		{version: "11.0.0", want: false},
		{version: "2.17.1", want: false},
		{version: "1.0.0-SNAPSHOT", want: false},
		{version: "r09", want: false},
		{version: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			if got := isPlaceholderVersion(tt.version); got != tt.want {
				t.Errorf("isPlaceholderVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

// TestCommitOnlyNoLocalPath reproduces the failure some clients (e.g. GitLab, which
// has no on-disk checkout) hit: passing a commit without a local path used to make
// osv-scanner fall back to scanning the filesystem root.
//
// Not run in parallel: osv-scanner's SetLogger writes to a shared, unsynchronized
// global, so running this concurrently with another test that also calls
// ListUnfixedVulnerabilities (e.g. TestEmptyProject) trips the race detector.
func TestCommitOnlyNoLocalPath(t *testing.T) {
	var client osvClient
	commit := "0000000000000000000000000000000000000000"
	_, err := client.ListUnfixedVulnerabilities(t.Context(), commit, "")
	if err != nil {
		t.Fatalf("commit without a local path shouldn't throw an error: %v", err)
	}
}
