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

package remediation

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	"github.com/ossf/scorecard/v4/rule"
)

func TestRepeatedSetup(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	for i := 0; i < 2; i++ {
		mockRepo := mockrepo.NewMockRepoClient(ctrl)
		mockRepo.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()
		uri := fmt.Sprintf("github.com/ossf/scorecard%d", i)
		mockRepo.EXPECT().URI().Return(uri).AnyTimes()

		c := checker.CheckRequest{
			RepoClient: mockRepo,
		}
		rmd, err := New(&c)
		if err != nil {
			t.Error(err)
		}

		want := fmt.Sprintf("ossf/scorecard%d", i)
		if rmd.Repo != want {
			t.Errorf("failed. expected: %v, got: %v", want, rmd.Repo)
		}
	}
}

func asPointer(s string) *string {
	return &s
}

type stubDigester struct{}

func (s stubDigester) Digest(name string) (string, error) {
	m := map[string]string{
		"foo":               "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
		"baz":               "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9",
		"amazoncorretto:11": "b1a711069b801a325a30885f08f5067b2b102232379750dda4d25a016afd9a88",
	}
	hash, ok := m[name]
	if !ok {
		//nolint:goerr113
		return "", fmt.Errorf("no hash for image: %q", name)
	}
	return fmt.Sprintf("sha256:%s", hash), nil
}

func TestCreateDockerfilePinningRemediation(t *testing.T) {
	t.Parallel()

	//nolint:govet,lll
	tests := []struct {
		name     string
		dep      checker.Dependency
		expected *rule.Remediation
	}{
		{
			name:     "no depdendency",
			dep:      checker.Dependency{},
			expected: nil,
		},
		{
			name: "image name no tag",
			dep: checker.Dependency{
				Name: asPointer("foo"),
				Type: checker.DependencyUseTypeDockerfileContainerImage,
			},
			expected: &rule.Remediation{
				Text:     "pin your Docker image by updating foo to foo@sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
				Markdown: "pin your Docker image by updating foo to foo@sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			},
		},
		{
			// github.com/ossf/scorecard/issues/2581
			name: "image name with tag",
			dep: checker.Dependency{
				Name:     asPointer("amazoncorretto"),
				PinnedAt: asPointer("11"),
				Type:     checker.DependencyUseTypeDockerfileContainerImage,
			},
			expected: &rule.Remediation{
				Text:     "pin your Docker image by updating amazoncorretto:11 to amazoncorretto:11@sha256:b1a711069b801a325a30885f08f5067b2b102232379750dda4d25a016afd9a88",
				Markdown: "pin your Docker image by updating amazoncorretto:11 to amazoncorretto:11@sha256:b1a711069b801a325a30885f08f5067b2b102232379750dda4d25a016afd9a88",
			},
		},
		{
			name: "unknown image",
			dep: checker.Dependency{
				Name: asPointer("not-found"),
				Type: checker.DependencyUseTypeDockerfileContainerImage,
			},
			expected: nil,
		},
		{
			name: "unknown tag",
			dep: checker.Dependency{
				Name:     asPointer("foo"),
				PinnedAt: asPointer("not-found"),
				Type:     checker.DependencyUseTypeDockerfileContainerImage,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CreateDockerfilePinningRemediation(&tt.dep, stubDigester{})
			if !cmp.Equal(got, tt.expected) {
				t.Errorf(cmp.Diff(got, tt.expected))
			}
		})
	}
}

func TestCreateWorkflowPinningRemediation(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name     string
		branch   string
		repo     string
		filepath string
		expected *rule.Remediation
	}{
		{
			name:     "valid input",
			branch:   "main",
			repo:     "ossf/scorecard",
			filepath: ".github/workflows/scorecard.yml",
			expected: &rule.Remediation{
				Text:     fmt.Sprintf(workflowText, "ossf/scorecard", "scorecard.yml", "main", "pin"),
				Markdown: fmt.Sprintf(workflowMarkdown, "ossf/scorecard", "scorecard.yml", "main", "pin"),
			},
		},
		{
			name:     "empty branch",
			branch:   "",
			repo:     "ossf/scorecard",
			filepath: ".github/workflows/<workflow-file>",
			expected: nil,
		},
		{
			name:     "empty repo",
			branch:   "main",
			repo:     "",
			filepath: ".github/workflows/<workflow-file>",
			expected: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := RemediationMetadata{
				Branch: tt.branch,
				Repo:   tt.repo,
			}
			got := r.CreateWorkflowPinningRemediation(tt.filepath)
			if !cmp.Equal(got, tt.expected) {
				t.Errorf(cmp.Diff(got, tt.expected))
			}
		})
	}
}
