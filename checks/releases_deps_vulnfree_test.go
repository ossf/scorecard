// Copyright 2025 OpenSSF Scorecard Authors
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

package checks

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestReleasesDirectDepsVulnFree(t *testing.T) {
	t.Parallel()

	//nolint:govet // Field alignment is a minor optimization
	tests := []struct {
		name          string
		releases      []clients.Release
		listErr       error
		wantErr       bool
		expectedScore int
	}{
		{
			name:          "no releases found",
			releases:      []clients.Release{},
			expectedScore: checker.InconclusiveResultScore,
		},
		{
			name:     "error listing releases",
			releases: nil,
			listErr:  errors.New("list error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListReleases().DoAndReturn(func() ([]clients.Release, error) {
				if tt.listErr != nil {
					return nil, tt.listErr
				}
				return tt.releases, nil
			}).AnyTimes()

			mockRepo.EXPECT().URI().Return("github.com/test/repo").AnyTimes()

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepo,
				Ctx:        t.Context(),
				Dlogger:    &dl,
			}

			res := ReleasesDirectDepsVulnFree(&req)

			if tt.wantErr {
				if res.Error == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if res.Error != nil {
				t.Errorf("unexpected error: %v", res.Error)
			}

			if res.Score != tt.expectedScore {
				t.Errorf("expected score %d, got %d", tt.expectedScore, res.Score)
			}
		})
	}
}

// TestReleasesDepsDebug tests the debug flag parsing.
// Note: This test cannot use t.Parallel() because it uses t.Setenv().
func TestReleasesDepsDebug(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{name: "1", envValue: "1", want: true},
		{name: "true", envValue: "true", want: true},
		{name: "TRUE", envValue: "TRUE", want: true},
		{name: "True", envValue: "True", want: true},
		{name: "yes", envValue: "yes", want: true},
		{name: "on", envValue: "on", want: true},
		{name: "ON", envValue: "ON", want: true},
		{name: "false", envValue: "false", want: false},
		{name: "0", envValue: "0", want: false},
		{name: "no", envValue: "no", want: false},
		{name: "off", envValue: "off", want: false},
		{name: "empty", envValue: "", want: false},
		{name: "random", envValue: "random", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() with t.Setenv()

			// Set test value using t.Setenv
			if tt.envValue != "" {
				t.Setenv("RELEASES_DEPS_DEBUG", tt.envValue)
			} else {
				t.Setenv("RELEASES_DEPS_DEBUG", "")
			}

			got := releasesDepsDebug()
			if got != tt.want {
				t.Errorf("releasesDepsDebug() with env=%q = %v, want %v", tt.envValue, got, tt.want)
			}
		})
	}
} // TestCheckReleasesDirectDepsVulnFreeRegistration verifies the check is properly registered.
func TestCheckReleasesDirectDepsVulnFreeRegistration(t *testing.T) {
	t.Parallel()

	// Verify the check name constant
	if CheckReleasesDirectDepsVulnFree != "ReleasesDirectDepsVulnFree" {
		t.Errorf("CheckReleasesDirectDepsVulnFree = %q, want \"ReleasesDirectDepsVulnFree\"",
			CheckReleasesDirectDepsVulnFree)
	}

	// Verify the function is callable (basic smoke test)
	// We can't easily test registration without accessing unexported functions,
	// but we can verify a basic error case works.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockRepoClient(ctrl)
	mockRepo.EXPECT().ListReleases().Return([]clients.Release{}, nil).AnyTimes()
	mockRepo.EXPECT().URI().Return("github.com/test/repo").AnyTimes()

	dl := scut.TestDetailLogger{}
	req := checker.CheckRequest{
		RepoClient: mockRepo,
		Ctx:        t.Context(),
		Dlogger:    &dl,
	}

	// Should return inconclusive for no releases
	res := ReleasesDirectDepsVulnFree(&req)
	if res.Score != checker.InconclusiveResultScore {
		t.Errorf("ReleasesDirectDepsVulnFree with no releases returned score %d, want %d",
			res.Score, checker.InconclusiveResultScore)
	}
}
