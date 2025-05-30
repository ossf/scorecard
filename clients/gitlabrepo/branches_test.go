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
	"sync"
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestGetBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		returnStatus        *gitlab.Response
		queryBranchReturn   *gitlab.Branch
		branchReturn        *gitlab.ProtectedBranch
		apprvlReturn        *gitlab.ProjectApprovals
		name                string
		branchName          string
		expectedBranchName  string
		projectID           string
		projectChecksReturn []*gitlab.ProjectStatusCheck
	}{
		{
			name:               "TargetCommittishFromRelease",
			branchName:         "/myproject/-/commit/magiccommitid",
			expectedBranchName: "",
			projectID:          "",
		},
		{
			name:               "Existing Protected Branch",
			branchName:         "branchName",
			expectedBranchName: "branchName",
			queryBranchReturn: &gitlab.Branch{
				Name:      "branchName",
				Protected: true,
			},
			returnStatus: &gitlab.Response{
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			},
			branchReturn:        &gitlab.ProtectedBranch{},
			projectChecksReturn: []*gitlab.ProjectStatusCheck{},
			apprvlReturn:        &gitlab.ProjectApprovals{},
		},
		{
			name:               "Existing UnProtected Branch",
			branchName:         "branchName",
			expectedBranchName: "branchName",
			queryBranchReturn: &gitlab.Branch{
				Name:      "branchName",
				Protected: false,
			},
			returnStatus: &gitlab.Response{
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := branchesHandler{
				once: new(sync.Once),
				repourl: &Repo{
					projectID: "5000",
				},
				queryBranch: func(pid interface{}, branch string,
					options ...gitlab.RequestOptionFunc,
				) (*gitlab.Branch, *gitlab.Response, error) {
					return tt.queryBranchReturn, tt.returnStatus, nil
				},
				getProtectedBranch: func(pid interface{}, branch string,
					options ...gitlab.RequestOptionFunc,
				) (*gitlab.ProtectedBranch, *gitlab.Response, error) {
					return tt.branchReturn, tt.returnStatus, nil
				},
				getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions,
					options ...gitlab.RequestOptionFunc,
				) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
					return tt.projectChecksReturn, tt.returnStatus, nil
				},
				getApprovalConfiguration: func(pid interface{}, options ...gitlab.RequestOptionFunc) (
					*gitlab.ProjectApprovals, *gitlab.Response, error,
				) {
					return tt.apprvlReturn, tt.returnStatus, nil
				},
			}

			handler.once.Do(func() {})

			br, err := handler.getBranch(tt.branchName)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}

			if *br.Name != tt.expectedBranchName {
				t.Errorf("Branch Name (%s) didn't match expected value (%s)", *br.Name, tt.expectedBranchName)
			}

			if tt.queryBranchReturn == nil {
				return
			}

			actualBool := *br.Protected
			expectedBool := tt.queryBranchReturn.Protected

			if actualBool != expectedBool {
				t.Errorf("Branch Protection didn't match expectation")
			}
		})
	}
}
