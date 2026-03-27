// Copyright 2021 OpenSSF Scorecard Authors
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

package clients

import "testing"

func TestHasMemoryUnsafeLanguage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		langs []Language
		want  bool
	}{
		{
			name:  "empty",
			langs: []Language{},
			want:  false,
		},
		{
			name: "safe",
			langs: []Language{
				{Name: Go, NumLines: 100},
				{Name: Python, NumLines: 100},
			},
			want: false,
		},
		{
			name: "contains C",
			langs: []Language{
				{Name: C, NumLines: 100},
			},
			want: true,
		},
		{
			name: "contains C++",
			langs: []Language{
				{Name: Cpp, NumLines: 100},
			},
			want: true,
		},
		{
			name: "contains Objective-C",
			langs: []Language{
				{Name: ObjectiveC, NumLines: 100},
			},
			want: true,
		},
		{
			name: "contains Objective-C++",
			langs: []Language{
				{Name: ObjectiveCpp, NumLines: 100},
			},
			want: true,
		},
		{
			name: "mixed safe and unsafe",
			langs: []Language{
				{Name: Go, NumLines: 100},
				{Name: C, NumLines: 100},
			},
			want: true,
		},
		{
			name: "cased variants",
			langs: []Language{
				{Name: LanguageName("C"), NumLines: 100},
			},
			want: true,
		},
		{
			name: "cased variants C++",
			langs: []Language{
				{Name: LanguageName("C++"), NumLines: 100},
			},
			want: true,
		},
		{
			name: "lowercase variant",
			langs: []Language{
				{Name: LanguageName("c"), NumLines: 100},
			},
			want: true,
		},
		{
			name: "mixed case variant",
			langs: []Language{
				{Name: LanguageName("C++"), NumLines: 100},
			},
			want: true,
		},
		{
			name: "lowercase objective-c",
			langs: []Language{
				{Name: LanguageName("objective-c"), NumLines: 100},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasMemoryUnsafeLanguage(tt.langs); got != tt.want {
				t.Errorf("HasMemoryUnsafeLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}
