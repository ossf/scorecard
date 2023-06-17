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

package githubrepo

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v38/github"

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
			responsePath: "./testdata/valid-webhook.json",
			want: []clients.Webhook{
				{
					ID:             12345678,
					UsesAuthSecret: false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			httpClient := &http.Client{
				Transport: stubTripper{
					responsePath: tt.responsePath,
				},
			}
			client := github.NewClient(httpClient)
			handler := &webhookHandler{
				ghClient: client,
				ctx:      ctx,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				repo:      "foo",
				commitSHA: clients.HeadSHA,
			}
			handler.init(ctx, &repoURL)
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
