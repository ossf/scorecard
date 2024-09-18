// Copyright 2023 OpenSSF Scorecard Authors
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
package checker

import (
	"testing"

	"github.com/ossf/scorecard/v5/finding"
)

func TestFile_Location(t *testing.T) {
	t.Parallel()
	file := File{
		Type:      finding.FileTypeSource,
		Path:      "bar.go",
		Offset:    10,
		EndOffset: 20,
		Snippet:   "some code",
	}

	loc := file.Location()

	if loc.Type != finding.FileTypeSource {
		t.Errorf("Expected loc.Type to be 'foo', got %v", loc.Type)
	}
	if loc.Path != "bar.go" {
		t.Errorf("Expected loc.Path to be 'bar.go', got %v", loc.Path)
	}
	if *loc.LineStart != 10 {
		t.Errorf("Expected *loc.LineStart to be 10, got %v", *loc.LineStart)
	}
	if *loc.LineEnd != 20 {
		t.Errorf("Expected *loc.LineEnd to be 20, got %v", *loc.LineEnd)
	}
	if *loc.Snippet != "some code" {
		t.Errorf("Expected *loc.Snippet to be 'some code', got %v", *loc.Snippet)
	}
}

func TestPinningDependenciesData_GetStagedDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     PinningDependenciesData
		useType  DependencyUseType
		expected []Dependency
	}{
		{
			name: "No staged dependencies",
			data: PinningDependenciesData{
				StagedDependencies: []Dependency{},
			},
			useType:  DependencyUseTypeGHAction,
			expected: []Dependency{},
		},
		{
			name: "Single matching dependency",
			data: PinningDependenciesData{
				StagedDependencies: []Dependency{
					{
						Name: newString("dep1"),
						Type: DependencyUseTypeGHAction,
					},
				},
			},
			useType: DependencyUseTypeGHAction,
			expected: []Dependency{
				{
					Name: newString("dep1"),
					Type: DependencyUseTypeGHAction,
				},
			},
		},
		{
			name: "Multiple dependencies with one match",
			data: PinningDependenciesData{
				StagedDependencies: []Dependency{
					{
						Name: newString("dep1"),
						Type: DependencyUseTypeGHAction,
					},
					{
						Name: newString("dep2"),
						Type: DependencyUseTypeDockerfileContainerImage,
					},
				},
			},
			useType: DependencyUseTypeGHAction,
			expected: []Dependency{
				{
					Name: newString("dep1"),
					Type: DependencyUseTypeGHAction,
				},
			},
		},
		{
			name: "Multiple dependencies with multiple matches",
			data: PinningDependenciesData{
				StagedDependencies: []Dependency{
					{
						Name: newString("dep1"),
						Type: DependencyUseTypeGHAction,
					},
					{
						Name: newString("dep2"),
						Type: DependencyUseTypeGHAction,
					},
				},
			},
			useType: DependencyUseTypeGHAction,
			expected: []Dependency{
				{
					Name: newString("dep1"),
					Type: DependencyUseTypeGHAction,
				},
				{
					Name: newString("dep2"),
					Type: DependencyUseTypeGHAction,
				},
			},
		},
		{
			name: "No matching dependencies",
			data: PinningDependenciesData{
				StagedDependencies: []Dependency{
					{
						Name: newString("dep1"),
						Type: DependencyUseTypeDockerfileContainerImage,
					},
				},
			},
			useType:  DependencyUseTypeGHAction,
			expected: []Dependency{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.data.GetStagedDependencies(tt.useType)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expected), len(result))
			}
			for i, dep := range result {
				if *dep.Name != *tt.expected[i].Name || dep.Type != tt.expected[i].Type {
					t.Errorf("Expected dependency %v, got %v", tt.expected[i], dep)
				}
			}
		})
	}
}

func newString(s string) *string {
	return &s
}
