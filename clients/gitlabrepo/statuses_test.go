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
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

func Test_listStatuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		responsePath string
		want         []clients.Status
		wantErr      bool
	}{
		{
			name:         "valid webhook",
			responsePath: "./testdata/valid-status",
			want: []clients.Status{
				{
					State:     "pending",
					Context:   "bundler:audit",
					URL:       "https://gitlab.example.com/janedoe/gitlab-foss/builds/91",
					TargetURL: "https://gitlab.example.com/janedoe/gitlab-foss/builds/91",
				},
			},
			wantErr: false,
		},
		{
			name:         "invalid webhook",
			responsePath: "./testdata/invalid-status",
			want:         nil,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpClient := &http.Client{
				Transport: stubTripper{
					responsePath: tt.responsePath,
				},
			}
			client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("gitlab.NewClient error: %v", err)
			}
			handler := &statusesHandler{
				glClient: client,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			got, err := handler.listStatuses("HEAD")
			if (err != nil) != tt.wantErr {
				t.Fatalf("listStatuses error: %v, wantedErr: %t", err, tt.wantErr)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("listStatuses() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}
