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

package raw

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestVulnerabilities(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name            string
		want            checker.VulnerabilitiesData
		err             error
		wantErr         bool
		vulnsResponse   clients.VulnerabilitiesResponse
		numberofCommits int
		expected        scut.TestReturn
		vulnsError      bool
	}{
		{
			name:            "Valid response",
			wantErr:         false,
			vulnsResponse:   clients.VulnerabilitiesResponse{},
			numberofCommits: 1,
		},
		{
			name:    "err response",
			wantErr: true,
			//nolint
			err:           errors.New("error"),
			vulnsResponse: clients.VulnerabilitiesResponse{},
		},
		{
			name:            "no commits",
			wantErr:         false,
			numberofCommits: 0,
			vulnsResponse:   clients.VulnerabilitiesResponse{},
		},
		{
			name:            "vulns err response",
			wantErr:         true,
			vulnsError:      true,
			numberofCommits: 1,
			vulnsResponse:   clients.VulnerabilitiesResponse{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.err != nil {
					return nil, tt.err
				}
				if tt.numberofCommits == 0 {
					return nil, nil
				}
				return []clients.Commit{{SHA: "test"}}, nil
			}).AnyTimes()

			mockRepo.EXPECT().LocalPath().DoAndReturn(func() (string, error) {
				return "test_path", nil
			}).AnyTimes()

			mockVulnClient := mockrepo.NewMockVulnerabilitiesClient(ctrl)
			mockVulnClient.EXPECT().HasUnfixedVulnerabilities(context.TODO(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, commit string, localPath string) (clients.VulnerabilitiesResponse, error) {
					if tt.vulnsError {
						//nolint
						return clients.VulnerabilitiesResponse{}, errors.New("error")
					}
					return tt.vulnsResponse, tt.err
				}).AnyTimes()

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient:            mockRepo,
				Ctx:                   context.TODO(),
				VulnerabilitiesClient: mockVulnClient,
				Dlogger:               &dl,
			}
			got, err := Vulnerabilities(&req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Vulnerabilities() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(got.Vulnerabilities) != len(tt.want.Vulnerabilities) {
					t.Errorf("Vulnerabilities() got = %v, want %v", len(got.Vulnerabilities), len(tt.want.Vulnerabilities))
				}
			}

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &checker.CheckResult{}, &dl) {
				t.Fatalf("Test %s failed", tt.name)
			}
		})
	}
}
