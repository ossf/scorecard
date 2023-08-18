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
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

func Test_CheckRuns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		responsePath string
		want         []clients.CheckRun
		wantErr      bool
	}{
		{
			name:         "valid checkruns",
			responsePath: "./testdata/valid-checkruns",
			want: []clients.CheckRun{
				{
					Status:     "pending",
					URL:        "https://example.com/foo/bar/pipelines/48",
					Conclusion: "",
				},
			},
			wantErr: false,
		},
		{
			name:         "valid checkruns with zero results",
			responsePath: "./testdata/valid-checkruns-1",
			wantErr:      false,
		},
		{
			name:         "failure fetching the checkruns",
			responsePath: "./testdata/invalid-checkruns-result",
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
			handler := &checkrunsHandler{
				glClient: client,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			got, err := handler.listCheckRunsForRef("main")
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkRuns error: %v, wantedErr: %t", err, tt.wantErr)
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("checkRuns() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}
