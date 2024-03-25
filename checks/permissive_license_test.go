// Copyright 2024 OpenSSF Scorecard Authors
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
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestPermissiveLicenseFileSubdirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		license  clients.License
		err      error
		expected scut.TestReturn
	}{
		{
			name: "With permissive LICENSE",
			license: clients.License{
				Key:    "Apache-2.0",
				Name:   "Apache 2.0",
				Path:   "testdata/licensedir/withpermissivelicense/LICENSE",
				SPDXId: "Apache-2.0",
				Type:   "Permissive",
				Size:   42,
			},
			expected: scut.TestReturn{
				Error:        nil,
				Score:        checker.MaxResultScore,
				NumberOfInfo: 0,
				NumberOfWarn: 0,
			},
			err: nil,
		},
		{
			name: "Without permissive LICENSE",
			license: clients.License{
				Key:    "AGPL-3.0",
				Name:   "AGPL 3.0",
				Path:   "testdata/licensedir/withcopyleftlicense/LICENSE",
				SPDXId: "AGPL-3.0",
				Type:   "Copyleft",
				Size:   42,
			},
			expected: scut.TestReturn{
				Error:        nil,
				Score:        checker.MinResultScore,
				NumberOfWarn: 0,
				NumberOfInfo: 1,
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// TODO: Use gMock instead of Localdir here.
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().ListLicenses().Return([]clients.License{tt.license}, nil).AnyTimes()

			ctx := context.Background()

			dl := scut.TestDetailLogger{}

			req := checker.CheckRequest{
				Ctx:        ctx,
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			res := PermissiveLicense(&req)

			scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl)

			ctrl.Finish()
		})
	}
}
