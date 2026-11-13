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

package raw

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

// attestationRepoClient is a RepoClient which can also list attestations.
type attestationRepoClient struct {
	clients.RepoClient
	listed             int
	listedAttestations int
}

func (c *attestationRepoClient) ListReleases() ([]clients.Release, error) {
	c.listed++
	return []clients.Release{{TagName: "listed"}}, nil
}

func (c *attestationRepoClient) ListReleasesWithAttestations() ([]clients.Release, error) {
	c.listedAttestations++
	return []clients.Release{{TagName: "attested"}}, nil
}

func TestSignedReleases_UsesAttestationLister(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	projectClient := mockrepo.NewMockProjectPackageClient(ctrl)
	projectClient.EXPECT().GetProjectPackageVersions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("not found")).AnyTimes()

	repoClient := &attestationRepoClient{}
	mockRepo := mockrepo.NewMockRepo(ctrl)
	mockRepo.EXPECT().Host().Return("github.com").AnyTimes()
	mockRepo.EXPECT().Path().Return("owner/repo").AnyTimes()

	got, err := SignedReleases(&checker.CheckRequest{
		RepoClient:    repoClient,
		ProjectClient: projectClient,
		Repo:          mockRepo,
	})
	if err != nil {
		t.Fatalf("SignedReleases: %v", err)
	}
	if repoClient.listed != 0 || repoClient.listedAttestations != 1 {
		t.Errorf("expected only ListReleasesWithAttestations to be called, got %d and %d",
			repoClient.listed, repoClient.listedAttestations)
	}
	if len(got.Releases) != 1 || got.Releases[0].TagName != "attested" {
		t.Errorf("unexpected releases: %+v", got.Releases)
	}
}

func TestSignedReleases_FallsBackToListReleases(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	projectClient := mockrepo.NewMockProjectPackageClient(ctrl)
	projectClient.EXPECT().GetProjectPackageVersions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("not found")).AnyTimes()
	mockRepo := mockrepo.NewMockRepo(ctrl)
	mockRepo.EXPECT().Host().Return("gitlab.com").AnyTimes()
	mockRepo.EXPECT().Path().Return("owner/repo").AnyTimes()

	repoClient := mockrepo.NewMockRepoClient(ctrl)
	repoClient.EXPECT().ListReleases().Return([]clients.Release{{TagName: "listed"}}, nil)

	got, err := SignedReleases(&checker.CheckRequest{
		RepoClient:    repoClient,
		ProjectClient: projectClient,
		Repo:          mockRepo,
	})
	if err != nil {
		t.Fatalf("SignedReleases: %v", err)
	}
	if len(got.Releases) != 1 || got.Releases[0].TagName != "listed" {
		t.Errorf("unexpected releases: %+v", got.Releases)
	}
}
