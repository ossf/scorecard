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

package ossfuzz

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ossf/scorecard/v4/clients"
)

func TestClient(t *testing.T) {
	tests := []struct {
		name       string
		project    string
		statusFile string
		wantHit    bool
		wantErr    bool
	}{
		{
			name:       "present project",
			project:    "github.com/ossf/scorecard",
			statusFile: "status.json",
			wantHit:    true,
			wantErr:    false,
		},
		{
			name:       "non existent project",
			project:    "github.com/not/here",
			statusFile: "status.json",
			wantHit:    false,
			wantErr:    false,
		},
		{
			name:       "non existent project which is a substring of a present project",
			project:    "github.com/ossf/score",
			statusFile: "status.json",
			wantHit:    false,
			wantErr:    false,
		},
		{
			name:       "non existent status file",
			project:    "github.com/ossf/scorecard",
			statusFile: "not_here.json",
			wantHit:    false,
			wantErr:    true,
		},
		{
			name:       "invalid status file",
			project:    "github.com/ossf/scorecard",
			statusFile: "invalid.json",
			wantHit:    false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			url := setupServer(t)
			statusURL := fmt.Sprintf("%s/%s", url, tt.statusFile)
			c := CreateOSSFuzzClient(context.Background(), statusURL)
			req := clients.SearchRequest{Query: tt.project}
			resp, err := c.Search(req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got err %v, wantedErr: %t", err, tt.wantErr)
			}
			if (resp.Hits > 0) != tt.wantHit {
				t.Errorf("wantHit: %t, got %d hits", tt.wantHit, resp.Hits)
			}
		})
	}
}

func setupServer(t *testing.T) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := os.ReadFile("testdata" + r.URL.Path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write(b)
	}))
	t.Cleanup(server.Close)
	return server.URL
}
