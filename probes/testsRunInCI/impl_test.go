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

//nolint:stylecheck
package testsRunInCI

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

const (
	// CheckCITests is the registered name for CITests.
	CheckCITests = "CI-Tests"
)

// Important: tests must include findings with values.
// Testing only for the outcome is insufficient, because the
// values of the findings are important to the probe.
func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		findings []*finding.Finding
		err      error
	}{
		{
			name: "Has 1 CIInfo which has a successful CheckRun.",
			raw: &checker.RawResults{
				CITestResults: checker.CITestData{
					CIInfo: []checker.RevisionCIInfo{
						{
							HeadSHA:           "HeadSHA",
							PullRequestNumber: 1,
							CheckRuns: []clients.CheckRun{
								{
									Status:     "completed",
									Conclusion: "success",
									App:        clients.CheckRunApp{Slug: "e2e"},
								},
							},
							Statuses: []clients.Status{
								{
									State:     "not successful",
									Context:   CheckCITests,
									TargetURL: "e2e",
								},
							},
						},
					},
				},
			},
			findings: []*finding.Finding{
				{
					Outcome:  finding.OutcomePositive,
					Probe:    Probe,
					Message:  "CI test found: pr: 1, context: e2e",
					Location: &finding.Location{Type: 4},
				},
			},
		},
		{
			name: "Has 1 CIInfo which has a successful Status.",
			raw: &checker.RawResults{
				CITestResults: checker.CITestData{
					CIInfo: []checker.RevisionCIInfo{
						{
							HeadSHA:           "HeadSHA",
							PullRequestNumber: 1,
							CheckRuns: []clients.CheckRun{
								{
									Status:     "incomplete",
									Conclusion: "not successful",
									App:        clients.CheckRunApp{Slug: "e2e"},
								},
							},
							Statuses: []clients.Status{
								{
									State:     "success",
									Context:   CheckCITests,
									TargetURL: "e2e",
								},
							},
						},
					},
				},
			},
			findings: []*finding.Finding{
				{
					Outcome:  finding.OutcomePositive,
					Probe:    Probe,
					Message:  "CI test found: pr: HeadSHA, context: CI-Tests",
					Location: &finding.Location{Type: 4},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(len(tt.findings), len(findings)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			for i := range tt.findings {
				outcome := &tt.findings[i]
				f := &findings[i]
				if diff := cmp.Diff(*outcome, f); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_isTest(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "appveyor",
			args: args{
				s: "appveyor",
			},
			want: true,
		},
		{
			name: "circleci",
			args: args{
				s: "circleci",
			},
			want: true,
		},
		{
			name: "jenkins",
			args: args{
				s: "jenkins",
			},
			want: true,
		},
		{
			name: "e2e",
			args: args{
				s: "e2e",
			},
			want: true,
		},
		{
			name: "github-actions",
			args: args{
				s: "github-actions",
			},
			want: true,
		},
		{
			name: "mergeable",
			args: args{
				s: "mergeable",
			},
			want: true,
		},
		{
			name: "packit-as-a-service",
			args: args{
				s: "packit-as-a-service",
			},
			want: true,
		},
		{
			name: "semaphoreci",
			args: args{
				s: "semaphoreci",
			},
			want: true,
		},
		{
			name: "test",
			args: args{
				s: "test",
			},
			want: true,
		},
		{
			name: "travis-ci",
			args: args{
				s: "travis-ci",
			},
			want: true,
		},
		{
			name: "azure-pipelines",
			args: args{
				s: "azure-pipelines",
			},
			want: true,
		},
		{
			name: "non-existing",
			args: args{
				s: "non-existing",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTest(tt.args.s); got != tt.want {
				t.Errorf("isTest() = %v, want %v for test %v", got, tt.want, tt.name)
			}
		})
	}
}

func Test_prHasSuccessfulCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    checker.RevisionCIInfo
		want    bool
		wantErr bool
	}{
		{
			name: "check run with conclusion success",
			args: checker.RevisionCIInfo{
				PullRequestNumber: 1,
				HeadSHA:           "sha",
				CheckRuns: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "success",
						URL:        "url",
						Status:     "completed",
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "check run with conclusion not success",
			args: checker.RevisionCIInfo{
				PullRequestNumber: 1,
				HeadSHA:           "sha",
				CheckRuns: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "failed",
						URL:        "url",
						Status:     "completed",
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "check run with conclusion not success",
			args: checker.RevisionCIInfo{
				PullRequestNumber: 1,
				HeadSHA:           "sha",
				CheckRuns: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "success",
						URL:        "url",
						Status:     "notcompleted",
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt

		//nolint:errcheck
		got, _, _ := prHasSuccessfulCheck(tt.args)
		if got != tt.want {
			t.Errorf("prHasSuccessfulCheck() = %v, want %v", got, tt.want)
		}
	}
}

func Test_prHasSuccessStatus(t *testing.T) {
	t.Parallel()
	type args struct {
		r checker.RevisionCIInfo
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "empty revision",
			args: args{
				r: checker.RevisionCIInfo{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "no statuses",
			args: args{
				r: checker.RevisionCIInfo{
					Statuses: []clients.Status{},
				},
			},
		},
		{
			name: "status is not success",
			args: args{
				r: checker.RevisionCIInfo{
					Statuses: []clients.Status{
						{
							State: "failure",
						},
					},
				},
			},
		},
		{
			name: "status is success",
			args: args{
				r: checker.RevisionCIInfo{
					Statuses: []clients.Status{
						{
							State:   "success",
							Context: CheckCITests,
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _, err := prHasSuccessStatus(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("prHasSuccessStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("prHasSuccessStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prHasSuccessfulCheckAdditional(t *testing.T) {
	t.Parallel()
	type args struct { //nolint:govet
		r  checker.RevisionCIInfo
		dl checker.DetailLogger
	}
	tests := []struct { //nolint:govet
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "empty revision",
			args: args{
				r: checker.RevisionCIInfo{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "status is not completed",
			args: args{
				r: checker.RevisionCIInfo{
					CheckRuns: []clients.CheckRun{
						{
							Status: "notcompleted",
						},
					},
				},
			},
		},
		{
			name: "status is not success",
			args: args{
				r: checker.RevisionCIInfo{
					CheckRuns: []clients.CheckRun{
						{
							Status:     "completed",
							Conclusion: "failure",
						},
					},
				},
			},
		},
		{
			name: "conclusion is success",
			args: args{
				r: checker.RevisionCIInfo{
					CheckRuns: []clients.CheckRun{
						{
							Status:     "completed",
							Conclusion: "success",
						},
					},
				},
			},
		},
		{
			name: "conclusion is success with a valid app slug",
			args: args{
				r: checker.RevisionCIInfo{
					CheckRuns: []clients.CheckRun{
						{
							Status:     "completed",
							Conclusion: "success",
							App:        clients.CheckRunApp{Slug: "e2e"},
						},
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _, err := prHasSuccessfulCheck(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("prHasSuccessfulCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("prHasSuccessfulCheck() got = %v, want %v", got, tt.want)
			}
		})
	}
}
