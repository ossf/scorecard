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

package raw

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

// TestCodeReviews tests the CodeReviews function.
func TestCodeReview(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		want    checker.CodeReviewData
		wantErr bool
	}{
		{
			name:    "Test_CodeReview",
			wantErr: false,
		},
		{
			name:    "Want error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mr := mockrepo.NewMockRepoClient(ctrl)
			mr.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.wantErr {
					//nolint
					return nil, errors.New("error")
				}
				return []clients.Commit{
					{
						SHA: "sha",
					},
				}, nil
			})
			result, err := CodeReview(mr)
			if (err != nil) != tt.wantErr {
				t.Errorf("CodeReview() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if !tt.wantErr && cmp.Equal(result, tt.want) {
				t.Errorf(cmp.Diff(result, tt.want))
			}
		})
	}
}
