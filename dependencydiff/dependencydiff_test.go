// Copyright 2022 Security Scorecard Authors
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

package dependencydiff

import (
	"context"
	"testing"

	"github.com/ossf/scorecard/v4/clients"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TestGetDependencyDiffResults is a test function for GetDependencyDiffResults
func TestGetDependencyDiffResults(t *testing.T) {
	t.Parallel()
	//nolint
	//nolint
	tests := []struct {
		name      string
		owner     string
		repo      string
		base      string
		head      string
		wantEmpty bool
		wantErr   bool
	}{
		{
			name:      "error response",
			owner:     "no_such_owner",
			repo:      "repo_not_exist",
			base:      clients.HeadSHA,
			head:      clients.HeadSHA,
			wantEmpty: true,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GetDependencyDiffResults(tt.owner, tt.repo, tt.base, tt.head)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDependencyDiffResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(got) == 0) != tt.wantEmpty {
				t.Errorf("FetchDependencyDiffData() = %v, want empty %v for %v", got, tt.wantEmpty, tt.name)
			}
		})
	}
}

// TestTestFetchDependencyDiffData is a test function for TestFetchDependencyDiffData
func TestFetchDependencyDiffData(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name      string
		ctx       context.Context
		owner     string
		repo      string
		base      string
		head      string
		wantEmpty bool
		wantErr   bool

		expected scut.TestReturn
	}{
		{
			name:      "error reponse",
			ctx:       context.Background(),
			owner:     "no_such_owner",
			repo:      "repo_not_exist",
			base:      clients.HeadSHA,
			head:      clients.HeadSHA,
			wantEmpty: true,
			wantErr:   true,
		},
	} // End of test cases.
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := FetchDependencyDiffData(tt.ctx, tt.owner, tt.repo, tt.base, tt.head)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchDependencyDiffData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(got) == 0) != tt.wantEmpty {
				t.Errorf("FetchDependencyDiffData() = %v, want empty %v for %v", got, tt.wantEmpty, tt.name)
			}
		})
	}
}
