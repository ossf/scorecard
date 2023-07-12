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

	"github.com/ossf/scorecard/v4/finding"
)

func TestFile_Location(t *testing.T) {
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
