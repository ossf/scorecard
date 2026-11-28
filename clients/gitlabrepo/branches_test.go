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
	"errors"
	"net/http"
	"sync"
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
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

func Test_branchesHandler_setup_nil_response(t *testing.T) {
	t.Parallel()

	errNetwork := errors.New("network timeout")

	//nolint:govet
	tests := []struct {
		name               string
		getProtectedBranch fnProtectedBranch
		getProjectChecks   fnListProjectStatusChecks
		getApprovalConfig  fnGetApprovalConfiguration
		wantErrIs          error
		expectErr          bool
	}{
		{
			name:      "nil response on getProtectedBranch",
			expectErr: true,
			wantErrIs: errNetwork,
			getProtectedBranch: func(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error) {
				return nil, nil, errNetwork
			},
		},
		{
			name:      "forbidden response on getProtectedBranch",
			expectErr: true,
			getProtectedBranch: func(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusForbidden}}, errors.New("forbidden")
			},
		},
		{
			name:      "nil response on getProjectChecks",
			expectErr: true,
			wantErrIs: errNetwork,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, nil, errNetwork
			},
		},
		{
			name:      "unauthorized response on getProjectChecks",
			expectErr: false,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusUnauthorized}}, errors.New("unauthorized")
			},
		},
		{
			name:      "forbidden response on getProjectChecks",
			expectErr: false,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusForbidden}}, errors.New("forbidden")
			},
		},
		{
			name:      "not found response on getProjectChecks",
			expectErr: false,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
			},
		},
		{
			name:      "server error response on getProjectChecks",
			expectErr: true,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, errors.New("server error")
			},
		},
		{
			name:      "nil response on getApprovalConfiguration",
			expectErr: true,
			wantErrIs: errNetwork,
			getApprovalConfig: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovals, *gitlab.Response, error) {
				return nil, nil, errNetwork
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
					commitSHA: clients.HeadSHA,
				},
				queryProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
					return &gitlab.Project{DefaultBranch: "main"}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
				queryBranch: func(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.Branch, *gitlab.Response, error) {
					return &gitlab.Branch{Name: "main", Protected: true}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
				getProtectedBranch: func(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error) {
					if tt.getProtectedBranch != nil {
						return tt.getProtectedBranch(pid, branch, options...)
					}
					return &gitlab.ProtectedBranch{Name: "main"}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
				getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
					if tt.getProjectChecks != nil {
						return tt.getProjectChecks(pid, opt, options...)
					}
					return []*gitlab.ProjectStatusCheck{}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
				getApprovalConfiguration: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovals, *gitlab.Response, error) {
					if tt.getApprovalConfig != nil {
						return tt.getApprovalConfig(pid, options...)
					}
					return &gitlab.ProjectApprovals{}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
			}

			// Should not panic on nil response and should match expected error status
			err := handler.setup()
			if (err != nil) != tt.expectErr {
				t.Errorf("handler.setup() error = %v, expectErr = %v", err, tt.expectErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("handler.setup() error = %v, want it to wrap %v", err, tt.wantErrIs)
			}
		})
	}
}

func Test_branchesHandler_getBranch_nil_response(t *testing.T) {
	t.Parallel()

	errNetwork := errors.New("network timeout")

	//nolint:govet
	tests := []struct {
		name              string
		getProjectChecks  fnListProjectStatusChecks
		getApprovalConfig fnGetApprovalConfiguration
		wantErrIs         error
		expectErr         bool
	}{
		{
			name:      "nil response on getProjectChecks in getBranch",
			expectErr: true,
			wantErrIs: errNetwork,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, nil, errNetwork
			},
		},
		{
			name:      "unauthorized response on getProjectChecks in getBranch",
			expectErr: false,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusUnauthorized}}, errors.New("unauthorized")
			},
		},
		{
			name:      "forbidden response on getProjectChecks in getBranch",
			expectErr: false,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusForbidden}}, errors.New("forbidden")
			},
		},
		{
			name:      "not found response on getProjectChecks in getBranch",
			expectErr: false,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
			},
		},
		{
			name:      "server error response on getProjectChecks in getBranch",
			expectErr: true,
			getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, errors.New("server error")
			},
		},
		{
			name:      "nil response on getApprovalConfiguration in getBranch",
			expectErr: true,
			wantErrIs: errNetwork,
			getApprovalConfig: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovals, *gitlab.Response, error) {
				return nil, nil, errNetwork
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
					commitSHA: clients.HeadSHA,
				},
				queryBranch: func(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.Branch, *gitlab.Response, error) {
					return &gitlab.Branch{Name: "main", Protected: true}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
				getProtectedBranch: func(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error) {
					return &gitlab.ProtectedBranch{Name: "main"}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
				getProjectChecks: func(pid interface{}, opt *gitlab.ListOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error) {
					if tt.getProjectChecks != nil {
						return tt.getProjectChecks(pid, opt, options...)
					}
					return []*gitlab.ProjectStatusCheck{}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
				getApprovalConfiguration: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovals, *gitlab.Response, error) {
					if tt.getApprovalConfig != nil {
						return tt.getApprovalConfig(pid, options...)
					}
					return &gitlab.ProjectApprovals{}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
				},
			}
			handler.once.Do(func() {})

			branch, err := handler.getBranch("main")
			if (err != nil) != tt.expectErr {
				t.Errorf("getBranch() error = %v, expectErr = %v", err, tt.expectErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("getBranch() error = %v, want it to wrap %v", err, tt.wantErrIs)
			}
			if !tt.expectErr && branch == nil {
				t.Errorf("expected non-nil branch when expectErr is false")
			}
		})
	}
}

func Test_branchesHandler_getBranch_missing_branch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		resp      *gitlab.Response
		name      string
		expectErr bool
	}{
		{
			name: "deleted branch is skipped",
			resp: &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}},
		},
		{
			name:      "server error is still returned",
			resp:      &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}},
			expectErr: true,
		},
		{
			name:      "nil response is still returned",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := branchesHandler{
				once: new(sync.Once),
				repourl: &Repo{
					projectID: "5000",
					commitSHA: clients.HeadSHA,
				},
				queryBranch: func(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.Branch, *gitlab.Response, error) {
					return nil, tt.resp, errors.New("query failed")
				},
			}
			handler.once.Do(func() {})

			branch, err := handler.getBranch("release-1.0")
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if branch != nil {
				t.Fatalf("expected no branch, got %v", branch)
			}
		})
	}
}
