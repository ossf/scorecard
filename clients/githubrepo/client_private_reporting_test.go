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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients"
)

type privateReportingRoundTripper struct {
	statusCode int
	body       string
	wantPath   string
}

func (rt privateReportingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		return nil, &unexpectedRequestError{got: req.Method, want: http.MethodGet}
	}
	if req.URL.Path != rt.wantPath {
		return nil, &unexpectedRequestError{got: req.URL.Path, want: rt.wantPath}
	}
	return &http.Response{
		StatusCode: rt.statusCode,
		Status:     http.StatusText(rt.statusCode),
		Body:       io.NopCloser(strings.NewReader(rt.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

type unexpectedRequestError struct {
	got  string
	want string
}

func (err *unexpectedRequestError) Error() string {
	return "unexpected request value: got " + err.got + ", want " + err.want
}

func TestIsPrivateVulnerabilityReportingEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want bool
	}{
		{name: "enabled", body: `{"enabled":true}`, want: true},
		{name: "disabled", body: `{"enabled":false}`, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rt := privateReportingRoundTripper{
				statusCode: http.StatusOK,
				body:       tc.body,
				wantPath:   "/repos/owner/repo/private-vulnerability-reporting",
			}
			client := &Client{
				ctx:        t.Context(),
				repourl:    &Repo{owner: "owner", repo: "repo", commitSHA: clients.HeadSHA},
				repoClient: github.NewClient(&http.Client{Transport: rt}),
			}

			got, err := client.IsPrivateVulnerabilityReportingEnabled()
			if err != nil {
				t.Fatalf("IsPrivateVulnerabilityReportingEnabled() error = %v", err)
			}
			if got != tc.want {
				t.Errorf("IsPrivateVulnerabilityReportingEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsPrivateVulnerabilityReportingEnabledRequiresHEAD(t *testing.T) {
	t.Parallel()
	client := &Client{repourl: &Repo{commitSHA: "a-commit-sha"}}

	_, err := client.IsPrivateVulnerabilityReportingEnabled()
	if !errors.Is(err, clients.ErrUnsupportedFeature) {
		t.Fatalf("IsPrivateVulnerabilityReportingEnabled() error = %v, want ErrUnsupportedFeature", err)
	}
}
