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

package raw

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

var tagv01 = "v01"

//nolint:govet
type tagArg struct {
	err    error
	name   string
	tagRef *clients.RepoRef
}

type tagsArg []tagArg

func (ba tagsArg) getTag(b string) (*clients.RepoRef, error) {
	for _, tag := range ba {
		if tag.name == b {
			return tag.tagRef, tag.err
		}
	}
	return nil, nil
}

func TestTagProtection(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name        string
		tags        tagsArg
		repoFiles   []string
		repoTags    []*clients.RepoRef
		releasesErr error
		want        checker.TagProtectionsData
		wantErr     error
	}{
		{
			name: "v01",
			repoTags: []*clients.RepoRef{
				{
					Name: &tagv01,
				},
			},
			tags: tagsArg{
				{
					name: "v01",
					tagRef: &clients.RepoRef{
						Name: &tagv01,
					},
				},
			},
			want: checker.TagProtectionsData{
				Tags: []clients.RepoRef{
					{
						Name: &tagv01,
					},
				},
				CodeownersFiles: []string{},
			},
			releasesErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().GetTag(gomock.Any()).AnyTimes().
				DoAndReturn(func(tag string) (*clients.RepoRef, error) {
					return tt.tags.getTag(tag)
				})
			mockRepoClient.EXPECT().ListTags().AnyTimes().
				DoAndReturn(func() ([]*clients.RepoRef, error) {
					return tt.repoTags, tt.releasesErr
				})
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).AnyTimes().Return(tt.repoFiles, nil)

			c := &checker.CheckRequest{
				RepoClient: mockRepoClient,
			}
			rawData, err := TagProtection(c)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("failed. expected: %v, got: %v", tt.wantErr, err)
				t.Fail()
			}
			if !cmp.Equal(rawData, tt.want) {
				t.Errorf("failed. expected: %v, got: %v", tt.want, rawData)
				t.Fail()
			}
		})
	}
}
