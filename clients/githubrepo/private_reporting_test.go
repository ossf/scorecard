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

package githubrepo

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v82/github"
)

func TestPrivateVulnerabilityReportingEnabled(t *testing.T) {
	t.Parallel()
	t.Run("uninitialized", func(t *testing.T) {
		t.Parallel()
		enabled, err := (&Client{}).PrivateVulnerabilityReportingEnabled()
		if enabled || !errors.Is(err, errRepoNotInitialized) {
			t.Fatalf("PrivateVulnerabilityReportingEnabled() = %v, %v", enabled, err)
		}
	})

	tests := []struct {
		name    string
		status  int
		body    string
		want    bool
		wantErr bool
	}{
		{name: "enabled", status: http.StatusOK, body: `{"enabled":true}`, want: true},
		{name: "disabled", status: http.StatusOK, body: `{"enabled":false}`, want: false},
		{name: "api error", status: http.StatusInternalServerError, body: `{"message":"nope"}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/ossf/scorecard/private-vulnerability-reporting" {
					t.Errorf("path = %s", r.URL.Path)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(srv.Close)

			base, err := url.Parse(srv.URL + "/")
			if err != nil {
				t.Fatalf("parse server url: %v", err)
			}
			gh := github.NewClient(srv.Client())
			gh.BaseURL = base
			gh.DisableRateLimitCheck = true
			client := &Client{
				ctx:        t.Context(),
				repoClient: gh,
				repourl:    &Repo{owner: "ossf", repo: "scorecard"},
			}
			enabled, err := client.PrivateVulnerabilityReportingEnabled()
			if tt.wantErr {
				if err == nil {
					t.Fatal("PrivateVulnerabilityReportingEnabled() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("PrivateVulnerabilityReportingEnabled() error = %v", err)
			}
			if enabled != tt.want {
				t.Fatalf("PrivateVulnerabilityReportingEnabled() = %v, want %v", enabled, tt.want)
			}
		})
	}
}
