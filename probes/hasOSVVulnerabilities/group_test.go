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

//nolint:stylecheck
package hasOSVVulnerabilities

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

func TestGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		vulns []clients.Vulnerability
		want  int
	}{
		{
			name: "alias matches ID",
			vulns: []clients.Vulnerability{
				{ID: "foo"},
				{ID: "bar", Aliases: []string{"foo"}},
			},
			want: 1,
		},
		{
			name: "no grouping",
			vulns: []clients.Vulnerability{
				{ID: "foo"},
				{ID: "bar"},
				{ID: "baz"},
			},
			want: 3,
		},
		{
			name: "transitive grouping",
			vulns: []clients.Vulnerability{
				{ID: "foo", Aliases: []string{"bar"}},
				{ID: "bar", Aliases: []string{"baz"}},
				{ID: "baz"},
			},
			want: 1,
		},
		{
			name: "multiple groups",
			vulns: []clients.Vulnerability{
				{ID: "foo", Aliases: []string{"bar"}},
				{ID: "bar"},
				{ID: "baz", Aliases: []string{"qux"}},
				{ID: "qux"},
			},
			want: 2,
		},
		{
			name:  "empty input",
			vulns: []clients.Vulnerability{},
			want:  0,
		},
		{
			name: "alias matches alias",
			vulns: []clients.Vulnerability{
				{ID: "foo", Aliases: []string{"bar"}},
				{ID: "baz", Aliases: []string{"bar"}},
			},
			want: 1,
		},
		{
			name: "transitive grouping with bridge last",
			vulns: []clients.Vulnerability{
				{ID: "foo", Aliases: []string{"bridge"}},
				{ID: "bar", Aliases: []string{"baz"}},
				{ID: "baz", Aliases: []string{"bridge"}},
			},
			want: 1,
		},
		{
			name: "duplicate IDs without aliases",
			vulns: []clients.Vulnerability{
				{ID: "foo"},
				{ID: "foo"},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			grouped := group(tt.vulns)
			if len(grouped) != tt.want {
				t.Errorf("expected %d vulns, got %d: %v", tt.want, len(grouped), grouped)
			}
		})
	}
}

func TestGroupAllPermutations(t *testing.T) {
	t.Parallel()

	input := []clients.Vulnerability{
		{ID: "foo", Aliases: []string{"link-a"}},
		{ID: "bar", Aliases: []string{"link-b"}},
		{ID: "baz", Aliases: []string{"link-b", "link-c"}},
		{ID: "qux", Aliases: []string{"link-a", "link-c"}},
	}
	want := []clients.Vulnerability{
		{
			ID: "bar",
			Aliases: []string{
				"bar", "baz", "foo",
				"link-a", "link-b", "link-c", "qux",
			},
		},
	}

	permutations := vulnerabilityPermutations(input)
	if len(permutations) != 24 {
		t.Fatalf("got %d permutations, want 24", len(permutations))
	}

	for i, vulns := range permutations {
		t.Run(fmt.Sprintf("order-%02d", i), func(t *testing.T) {
			t.Parallel()

			if got := group(vulns); !reflect.DeepEqual(got, want) {
				t.Fatalf("group() = %#v, want %#v", got, want)
			}

			raw := &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: vulns,
				},
			}
			findings, probe, err := Run(raw)
			if err != nil {
				t.Fatalf("Run() error: %v", err)
			}
			if probe != Probe {
				t.Errorf("probe = %q, want %q", probe, Probe)
			}
			if len(findings) != 1 {
				t.Fatalf("got %d findings, want 1", len(findings))
			}
			if got := findings[0].Message; got !=
				"Project is vulnerable to: https://osv.dev/bar" {
				t.Errorf("unexpected finding message: %q", got)
			}
		})
	}
}

func TestGroupContents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		vulns []clients.Vulnerability
		want  []clients.Vulnerability
	}{
		{
			name: "representative ID priority and alias deduplication",
			vulns: []clients.Vulnerability{
				{ID: "GHSA-test", Aliases: []string{"shared", "shared"}},
				{ID: "OTHER-test", Aliases: []string{"shared"}},
				{ID: "CVE-2025-0002", Aliases: []string{"shared"}},
				{ID: "DSA-0002", Aliases: []string{"shared"}},
				{ID: "DSA-0001", Aliases: []string{"shared"}},
			},
			want: []clients.Vulnerability{
				{
					ID: "DSA-0001",
					Aliases: []string{
						"CVE-2025-0002", "DSA-0001", "DSA-0002",
						"GHSA-test", "OTHER-test", "shared",
					},
				},
			},
		},
		{
			name: "independent groups remain separate",
			vulns: []clients.Vulnerability{
				{ID: "foo", Aliases: []string{"bar"}},
				{ID: "bar"},
				{ID: "baz", Aliases: []string{"qux"}},
				{ID: "qux"},
				{ID: "solo"},
			},
			want: []clients.Vulnerability{
				{ID: "bar", Aliases: []string{"bar", "foo"}},
				{ID: "baz", Aliases: []string{"baz", "qux"}},
				{ID: "solo", Aliases: []string{"solo"}},
			},
		},
		{
			name: "duplicate IDs preserve aliases",
			vulns: []clients.Vulnerability{
				{ID: "foo", Aliases: []string{"a"}},
				{ID: "foo", Aliases: []string{"b"}},
			},
			want: []clients.Vulnerability{
				{ID: "foo", Aliases: []string{"a", "b", "foo"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := group(tt.vulns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("group() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func vulnerabilityPermutations(input []clients.Vulnerability) [][]clients.Vulnerability {
	var result [][]clients.Vulnerability
	current := append([]clients.Vulnerability(nil), input...)

	var generate func(int)
	generate = func(start int) {
		if start == len(current) {
			result = append(result,
				append([]clients.Vulnerability(nil), current...))
			return
		}

		for i := start; i < len(current); i++ {
			current[start], current[i] = current[i], current[start]
			generate(start + 1)
			current[start], current[i] = current[i], current[start]
		}
	}

	generate(0)
	return result
}
