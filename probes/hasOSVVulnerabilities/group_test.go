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
	"testing"

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
