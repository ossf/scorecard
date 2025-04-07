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
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_CheckRuns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		responsePath string
		ref          string
		want         []clients.CheckRun
		wantErr      bool
	}{
		{
			name:         "valid checkruns",
			responsePath: "./testdata/valid-checkruns",
			ref:          "main",
			want: []clients.CheckRun{
				{
					Status:     "queued",
					URL:        "https://example.com/foo/bar/pipelines/48",
					Conclusion: "",
				},
			},
			wantErr: false,
		},
		{
			name:         "valid checkruns with zero results",
			responsePath: "./testdata/empty-response",
			ref:          "eb94b618fb5865b26e80fdd8ae531b7a63ad851a",
			wantErr:      false,
		},
		{
			name:         "failure fetching the checkruns",
			responsePath: "./testdata/invalid-checkruns-result",
			ref:          "main",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
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

			repoURL := Repo{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			got, err := handler.listCheckRunsForRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkRuns error: %v, wantedErr: %t", err, tt.wantErr)
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("checkRuns() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestParseGitlabStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status string
		want   clients.CheckRun
	}{
		{
			status: "created",
			want: clients.CheckRun{
				Status:     "queued",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "",
			},
		},
		{
			status: "waiting_for_resource",
			want: clients.CheckRun{
				Status:     "queued",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "",
			},
		},
		{
			status: "preparing",
			want: clients.CheckRun{
				Status:     "queued",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "",
			},
		},
		{
			status: "pending",
			want: clients.CheckRun{
				Status:     "queued",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "",
			},
		},
		{
			status: "scheduled",
			want: clients.CheckRun{
				Status:     "queued",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "",
			},
		},
		{
			status: "running",
			want: clients.CheckRun{
				Status:     "in_progress",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "",
			},
		},
		{
			status: "failed",
			want: clients.CheckRun{
				Status:     "completed",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "failure",
			},
		},
		{
			status: "success",
			want: clients.CheckRun{
				Status:     "completed",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "success",
			},
		},
		{
			status: "canceled",
			want: clients.CheckRun{
				Status:     "completed",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "cancelled",
			},
		},
		{
			status: "skipped",
			want: clients.CheckRun{
				Status:     "completed",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "skipped",
			},
		},
		{
			status: "manual",
			want: clients.CheckRun{
				Status:     "completed",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "action_required",
			},
		},
		{
			status: "invalid_status",
			want: clients.CheckRun{
				Status:     "invalid_status",
				URL:        "https://example.com/foo/bar/pipelines/48",
				Conclusion: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()

			info := gitlab.PipelineInfo{
				WebURL: "https://example.com/foo/bar/pipelines/48",
				Status: tt.status,
			}

			got := parseGitlabStatus(&info)

			if !cmp.Equal(got, tt.want) {
				t.Errorf("parseGitlabStatus() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}
