// Copyright 2026 OpenSSF Scorecard Authors
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

package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBadgeByProjectID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		serverResp  string
		wantLevel   string
		wantRepoURL string
		id          int
		statusCode  int
		wantErr     bool
	}{
		{
			name:        "gold badge with repo_url",
			id:          1234,
			serverResp:  `{"id":1234,"badge_level":"gold","repo_url":"https://github.com/ossf/scorecard"}`,
			statusCode:  http.StatusOK,
			wantLevel:   "gold",
			wantRepoURL: "https://github.com/ossf/scorecard",
		},
		{
			name:       "not found returns empty no error",
			id:         9999,
			serverResp: "",
			statusCode: http.StatusNotFound,
			wantLevel:  "",
		},
		{
			name:       "malformed JSON returns error",
			id:         1,
			serverResp: `{bad json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.serverResp != "" {
					w.Write([]byte(tt.serverResp)) //nolint:errcheck
				}
			}))
			defer srv.Close()

			client := &httpClientCIIBestPractices{baseURL: srv.URL}
			got, err := client.GetBadgeByProjectID(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetBadgeByProjectID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got.BadgeLevel != tt.wantLevel {
				t.Errorf("BadgeLevel = %q, want %q", got.BadgeLevel, tt.wantLevel)
			}
			if got.RepoURL != tt.wantRepoURL {
				t.Errorf("RepoURL = %q, want %q", got.RepoURL, tt.wantRepoURL)
			}
		})
	}
}

// TestCIIProjectClientInterface verifies that httpClientCIIBestPractices
// satisfies the optional CIIProjectClient interface at compile time.
func TestCIIProjectClientInterface(t *testing.T) {
	t.Parallel()
	var _ CIIProjectClient = &httpClientCIIBestPractices{}
}
