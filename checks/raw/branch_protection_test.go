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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

var (
	errBPTest         = errors.New("test error")
	defaultBranchName = "default"
	releaseBranchName = "release-branch"
	mainBranchName    = "main"
)

//nolint:govet
type branchArg struct {
	err           error
	name          string
	branchRef     *clients.BranchRef
	defaultBranch bool
}

type branchesArg []branchArg

func (ba branchesArg) getDefaultBranch() (*clients.BranchRef, error) {
	for _, branch := range ba {
		if branch.defaultBranch {
			return branch.branchRef, branch.err
		}
	}
	return nil, nil
}

func (ba branchesArg) getBranch(b string) (*clients.BranchRef, error) {
	for _, branch := range ba {
		if branch.name == b {
			return branch.branchRef, branch.err
		}
	}
	return nil, nil
}

func TestBranchProtection(t *testing.T) {
	t.Parallel()
	//nolint: govet
	tests := []struct {
		name        string
		branches    branchesArg
		repoFiles   []string
		releases    []clients.Release
		releasesErr error
		want        checker.BranchProtectionsData
		wantErr     error
	}{
		{
			name: "default-branch-err",
			branches: branchesArg{
				{
					name: defaultBranchName,
					err:  errBPTest,
				},
			},
			want: checker.BranchProtectionsData{
				CodeownersFiles: []string{},
			},
		},
		{
			name: "null-default-branch-only",
			branches: branchesArg{
				{
					name:          defaultBranchName,
					defaultBranch: true,
					branchRef:     nil,
				},
			},
			want: checker.BranchProtectionsData{
				CodeownersFiles: []string{},
			},
		},
		{
			name: "default-branch-only",
			branches: branchesArg{
				{
					name:          defaultBranchName,
					defaultBranch: true,
					branchRef: &clients.BranchRef{
						Name: &defaultBranchName,
					},
				},
			},
			want: checker.BranchProtectionsData{
				Branches: []clients.BranchRef{
					{
						Name: &defaultBranchName,
					},
				},
				CodeownersFiles: []string{},
			},
		},
		{
			name:        "list-releases-error",
			releasesErr: errBPTest,
			wantErr:     errBPTest,
		},
		{
			name: "no-releases",
			want: checker.BranchProtectionsData{
				CodeownersFiles: []string{},
			},
		},
		{
			name: "empty-targetcommitish",
			releases: []clients.Release{
				{
					TargetCommitish: "",
				},
			},
			wantErr: errInternalCommitishNil,
		},
		{
			name: "release-branch-err",
			releases: []clients.Release{
				{
					TargetCommitish: releaseBranchName,
				},
			},
			branches: branchesArg{
				{
					name: releaseBranchName,
					err:  errBPTest,
				},
			},
			wantErr: errBPTest,
		},
		{
			name: "nil-release-branch",
			releases: []clients.Release{
				{
					TargetCommitish: releaseBranchName,
				},
			},
			branches: branchesArg{
				{
					name:      releaseBranchName,
					branchRef: nil,
				},
			},
			want: checker.BranchProtectionsData{
				CodeownersFiles: []string{},
			},
		},
		{
			name: "add-release-branch",
			releases: []clients.Release{
				{
					TargetCommitish: releaseBranchName,
				},
			},
			branches: branchesArg{
				{
					name: releaseBranchName,
					branchRef: &clients.BranchRef{
						Name: &releaseBranchName,
					},
				},
			},
			want: checker.BranchProtectionsData{
				Branches: []clients.BranchRef{
					{
						Name: &releaseBranchName,
					},
				},
				CodeownersFiles: []string{},
			},
		},
		{
			name: "master-to-main-redirect",
			releases: []clients.Release{
				{
					TargetCommitish: "master",
				},
			},
			branches: branchesArg{
				{
					name: mainBranchName,
					branchRef: &clients.BranchRef{
						Name: &mainBranchName,
					},
				},
			},
			want: checker.BranchProtectionsData{
				Branches: []clients.BranchRef{
					{
						Name: &mainBranchName,
					},
				},
				CodeownersFiles: []string{},
			},
		},
		{
			name: "default-and-release-branches",
			releases: []clients.Release{
				{
					TargetCommitish: releaseBranchName,
				},
			},
			branches: branchesArg{
				{
					name:          defaultBranchName,
					defaultBranch: true,
					branchRef: &clients.BranchRef{
						Name: &defaultBranchName,
					},
				},
				{
					name: releaseBranchName,
					branchRef: &clients.BranchRef{
						Name: &releaseBranchName,
					},
				},
			},
			want: checker.BranchProtectionsData{
				Branches: []clients.BranchRef{
					{
						Name: &defaultBranchName,
					},
					{
						Name: &releaseBranchName,
					},
				},
				CodeownersFiles: []string{},
			},
		},
		// TODO: Add tests for commitSHA regex matching.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().GetDefaultBranch().
				AnyTimes().DoAndReturn(func() (*clients.BranchRef, error) {
				return tt.branches.getDefaultBranch()
			})
			mockRepoClient.EXPECT().GetBranch(gomock.Any()).AnyTimes().
				DoAndReturn(func(branch string) (*clients.BranchRef, error) {
					return tt.branches.getBranch(branch)
				})
			mockRepoClient.EXPECT().ListReleases().AnyTimes().
				DoAndReturn(func() ([]clients.Release, error) {
					return tt.releases, tt.releasesErr
				})
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).AnyTimes().Return(tt.repoFiles, nil)
			rawData, err := BranchProtection(mockRepoClient)
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
