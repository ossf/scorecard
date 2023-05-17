// Copyright 2022 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/xanzy/go-gitlab"

	moq "github.com/ossf/scorecard/v4/clients/gitlabrepo/moqs"
)

func TestCommitters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		commits      []*gitlab.Commit
		users        []*gitlab.User
		mergeRequest []*gitlab.MergeRequest
	}{
		{
			name: "Repeated Users",
			commits: []*gitlab.Commit{
				{
					CommitterName: "John Doe",
					AuthorEmail:   "noone@nowhere.com",
					CommittedDate: gitlab.Time(time.Now()),
				},
			},
			users: []*gitlab.User{
				{
					ID:           50,
					Organization: "",
					Bot:          false,
				},
			},
			mergeRequest: []*gitlab.MergeRequest{
				{
					ID:        51,
					CreatedAt: gitlab.Time(time.Now().AddDate(0, -1, 0)),
					MergedAt:  gitlab.Time(time.Now()),
					Author: &gitlab.BasicUser{
						ID: 52,
					},
					MergedBy: &gitlab.BasicUser{
						ID: 53,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			moqMethods := moq.NewMockTestMethods(ctrl)

			handler := commitsHandler{
				once:       new(sync.Once),
				moqMethods: moqMethods,
			}

			moqMethods.EXPECT().ListCommits().Return(tt.commits)
			moqMethods.EXPECT().QueryUsers("John Doe").Return(tt.users)
			moqMethods.EXPECT().ListMergeRequestsByCommit().Return(tt.mergeRequest)

			err := handler.setup()
			if err != nil {
				t.Errorf("Exception thrown %v", err)
			}
		})
	}
}

func TestParsingEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Perfect Match Email Parser",
			email:    "john.doe@nowhere.com",
			expected: "john doe",
		},
		{
			name:     "Valid Email Not Formatted as expected",
			email:    "johndoe@nowhere.com",
			expected: "johndoe@nowhere com",
		},
		{
			name:     "Invalid email format",
			email:    "johndoe@nowherecom",
			expected: "johndoe@nowherecom",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseEmailToName(tt.email)

			if tt.expected != result {
				t.Errorf("Parser didn't work as expected: %s != %s", tt.expected, result)
			}
		})
	}
}
