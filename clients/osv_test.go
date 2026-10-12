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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	transitiverequirements "github.com/google/osv-scalibr/enricher/transitivedependency/requirements"
	"github.com/google/osv-scanner/v2/pkg/osvscanner"
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

	actions, err := (osvClient{}).scannerActions(nil, nil)
	if err != nil {
		t.Fatalf("scannerActions failed: %v", err)
	}
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

	actions, err := (osvClient{local: true}).scannerActions(nil, nil)
	if err != nil {
		t.Fatalf("scannerActions failed: %v", err)
	}
	if !actions.TransitiveScanning.Disabled {
		t.Fatal("local client must disable transitive scanning to stay offline")
	}
}

type osvTestTransport func(*http.Request) (*http.Response, error)

func (f osvTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestScannerExcludesTestData(t *testing.T) {
	t.Parallel()

	for _, directory := range []string{
		"testdata",
		"testData",
		"TESTDATA",
		"nested/testData",
	} {
		for _, mode := range []string{"absolute", "relative"} {
			t.Run(directory+"/"+mode, func(t *testing.T) {
				t.Parallel()

				root := createOSVTestRepository(t, directory)
				scanPath := root
				if mode == "relative" {
					workingDirectory, err := os.Getwd()
					if err != nil {
						t.Fatal(err)
					}
					scanPath, err = filepath.Rel(workingDirectory, root)
					if err != nil {
						t.Skipf("relative path unavailable across filesystem volumes: %v", err)
					}
				}

				checkOSVTestDataExclusion(t, scanPath)
			})
		}
	}
}

func createOSVTestRepository(t *testing.T, directory string) string {
	t.Helper()

	root := filepath.Join(t.TempDir(), "testdata", "repository[1]+")
	files := map[string]string{
		directory + "/requirements.txt":   "fixture-package",
		"production/requirements.txt":     "production-package",
		"testData-extra/requirements.txt": "similar-directory-package",
		"mytestData/requirements.txt":     "prefix-directory-package",
		"testData.txt/requirements.txt":   "suffix-directory-package",
	}

	for relativePath, packageName := range files {
		filename := filepath.Join(root, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		content := []byte(packageName + "==1.0.0\n")
		if err := os.WriteFile(filename, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func checkOSVTestDataExclusion(t *testing.T, scanPath string) {
	t.Helper()

	actions, err := (osvClient{}).scannerActions([]string{scanPath}, nil)
	if err != nil {
		t.Fatalf("scannerActions failed: %v", err)
	}
	actions.TransitiveScanning.Disabled = true
	httpClient, querySnapshot := newOSVQueryRecorder()
	actions.HTTPClient = httpClient

	if _, err := osvscanner.DoScan(actions); err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	queried := querySnapshot()
	if queried["fixture-package"] {
		t.Error("test-data package was queried")
	}
	for _, name := range []string{
		"production-package",
		"similar-directory-package",
		"prefix-directory-package",
		"suffix-directory-package",
	} {
		if !queried[name] {
			t.Errorf("package %q should still be queried", name)
		}
	}
}

func newOSVQueryRecorder() (*http.Client, func() map[string]bool) {
	var mu sync.Mutex
	queried := map[string]bool{}

	client := &http.Client{
		Transport: osvTestTransport(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost || req.URL.Path != "/v1/querybatch" {
				return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL)
			}

			var batch struct {
				Queries []struct {
					Package struct {
						Name string `json:"name"`
					} `json:"package"`
				} `json:"queries"`
			}
			if err := json.NewDecoder(req.Body).Decode(&batch); err != nil {
				return nil, fmt.Errorf("decode OSV query: %w", err)
			}

			mu.Lock()
			for _, query := range batch.Queries {
				queried[query.Package.Name] = true
			}
			mu.Unlock()

			body, err := json.Marshal(struct {
				Results []struct{} `json:"results"`
			}{
				Results: make([]struct{}, len(batch.Queries)),
			})
			if err != nil {
				return nil, fmt.Errorf("encode OSV response: %w", err)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(string(body))),
				Request:    req,
			}, nil
		}),
	}

	snapshot := func() map[string]bool {
		mu.Lock()
		defer mu.Unlock()

		result := make(map[string]bool, len(queried))
		for name, present := range queried {
			result[name] = present
		}
		return result
	}
	return client, snapshot
}
