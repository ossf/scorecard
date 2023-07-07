// Copyright 2023 OpenSSF Scorecard Authors
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
	"strconv"
	"sync"
	"testing"

	"github.com/xanzy/go-gitlab"
)

func TestContributors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		contributors []*gitlab.Contributor
		users        []*gitlab.User
	}{
		{
			name:         "No Data",
			contributors: []*gitlab.Contributor{},
			users:        []*gitlab.User{},
		},
		{
			name: "Simple Passthru",
			contributors: []*gitlab.Contributor{
				{
					Name: "John Doe",
				},
			},
			users: []*gitlab.User{
				{
					Name: "John Doe",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := contributorsHandler{
				fnContributors: func(s string) ([]*gitlab.Contributor, error) {
					return tt.contributors, nil
				},
				fnUsers: func(s string) ([]*gitlab.User, error) {
					return tt.users, nil
				},
				once: new(sync.Once),
				repourl: &repoURL{
					commitSHA: "HEAD",
				},
			}

			if len(handler.contributors) != 0 {
				t.Errorf("Initial count of contributors should be 0, but was %v", strconv.Itoa(len(handler.contributors)))
			}

			err := handler.setup()
			if err != nil {
				t.Errorf("Exception in contributors.setup %v", err)
			}

			if len(handler.contributors) != len(tt.contributors) {
				t.Errorf("Initial count of contributors should be 1, but was %v", strconv.Itoa(len(handler.contributors)))
			}
		})
	}
}
