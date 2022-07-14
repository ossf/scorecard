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
	"github.com/ossf/scorecard/v4/log"
)

// TestGetDependencyDiffResults is a test function for GetDependencyDiffResults.
func TestGetDependencyDiffResults(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name             string
		owner            string
		repo             string
		base             string
		head             string
		ctx              context.Context
		logger           *log.Logger
		wantEmptyResults bool
		wantErr          bool
	}{
		{
			name:             "error response",
			owner:            "no_such_owner",
			repo:             "repo_not_exist",
			base:             "",
			head:             clients.HeadSHA,
			ctx:              context.Background(),
			logger:           log.NewLogger(log.InfoLevel),
			wantEmptyResults: true,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GetDependencyDiffResults(tt.ctx, tt.owner, tt.repo, tt.base, tt.head, nil, tt.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDependencyDiffResults() error = [%v], errWanted [%v]", err, tt.wantErr)
				return
			}
			if (len(got) == 0) != tt.wantEmptyResults {
				t.Errorf("GetDependencyDiffResults() = %v, want empty [%v] for [%v]", got, tt.wantEmptyResults, tt.name)
			}
		})
	}
}
