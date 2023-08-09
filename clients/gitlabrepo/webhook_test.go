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
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type stubTripper struct {
	responsePath string
}

func (s stubTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	f, err := os.Open(s.responsePath)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       f,
	}, nil
}

func Test_listWebhooks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		responsePath string
		want         []clients.Webhook
		wantErr      bool
	}{
		{
			name:         "valid webhook",
			responsePath: "./testdata/valid-webhook",
			want: []clients.Webhook{
				{
					ID:             1,
					Path:           "http://example.com/hook",
					UsesAuthSecret: true,
				},
			},
			wantErr: false,
		},
		{
			name:         "invalid webhook",
			responsePath: "./testdata/invalid-webhook",
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
			handler := &webhookHandler{
				glClient: client,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			got, err := handler.listWebhooks()
			if (err != nil) != tt.wantErr {
				t.Fatalf("listWebhooks error: %v, wantedErr: %t", err, tt.wantErr)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("listWebhooks() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}
