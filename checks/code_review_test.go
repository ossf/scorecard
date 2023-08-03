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

package checks

import (
	"errors"
	"testing"
	"time"
	
	"github.com/golang/mock/gomock"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes"
	"github.com/ossf/scorecard/v4/probes/zrunner"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

var (
	errNew = errors.New("error")
	errProbeReturned = errors.New("probe run failure")
)

// TestCodeReview tests the code review checker.
func TestCodereview(t *testing.T) {
	t.Parallel()
	//fieldalignment lint issue. Ignoring it as it is not important for this test.
	//nolint

	tests := []struct {
		err       error
		name      string
		commiterr error
		commits   []clients.Commit
		expected  checker.CheckResult
	}{
		{
			name: "no commits",
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name:      "no commits with error",
			commiterr: errNew,
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "no PR's with error",
			err:  errNew,
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name:      "no PR's with error as well as commits",
			err:       errNew,
			commiterr: errNew,
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "Valid GitHub PR",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						Reviews: []clients.Review{
							{
								Author: &clients.User{Login: "alice"},
								State:  "APPROVED",
							},
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Valid Prow PR as not a bot",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "user",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						Labels: []clients.Label{
							{
								Name: "lgtm",
							},
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Valid Prow PR and commits as bot",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						Labels: []clients.Label{
							{
								Name: "lgtm",
							},
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Valid PR's and commits with merged by someone else",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
						Labels: []clients.Label{
							{
								Name: "lgtm",
							},
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "2 PRs 1 review on GitHub",
			commits: []clients.Commit{
				{
					SHA: "a",
					Committer: clients.User{
						Login: "bob",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						Reviews: []clients.Review{
							{
								Author: &clients.User{Login: "alice"},
								State:  "APPROVED",
							},
						},
					},
				},
				{
					SHA: "sha2",
					Committer: clients.User{
						Login: "bob",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 5,
			},
		},
		{
			name: "implicit maintainer approval through merge",
			commits: []clients.Commit{
				{
					SHA: "abc",
					Committer: clients.User{
						Login: "bob",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						MergedBy: clients.User{Login: "alice"},
					},
				},
				{
					SHA: "def",
					Committer: clients.User{
						Login: "bob",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 5,
			},
		},
		{
			name: "Valid Phabricator commit",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					Message: "Title\nReviewed By: alice\nDifferential Revision: PHAB234",
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Phabricator like, missing differential",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					Message: "Title\nReviewed By: alice",
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "Phabricator like, missing reviewed by",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					Message: "Title\nDifferential Revision: PHAB234",
				},
			},
			expected: checker.CheckResult{
				Score: checker.MaxResultScore,
			},
		},
		{
			name: "Valid piper commit",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "",
					},
					Message: "Title\nPiperOrigin-RevId: 444529962",
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListCommits().Return(tt.commits, tt.err).AnyTimes()

			req := checker.CheckRequest{
				RepoClient: mockRepo,
			}
			req.Dlogger = &scut.TestDetailLogger{}
			res := CodeReview(&req)

			if tt.err != nil {
				if res.Error == nil {
					t.Errorf("Expected error %v, got nil", tt.err)
				}
				// return as we don't need to check the rest of the fields.
				return
			}

			if res.Score != tt.expected.Score {
				t.Errorf("Expected score %d, got %d for %v", tt.expected.Score, res.Score, tt.name)
			}
			ctrl.Finish()
		})
	}
}

// TestProbesCodereview tests the probes associated with the Code-Review check.
func TestProbesCodereview(t *testing.T) {
	t.Parallel()
	probeTests := []struct {
		name             string
		rawResults       *checker.RawResults
		err              error
		expectedFindings []finding.Finding
	}{
		{
			name: "no changesets",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{},
				},
			},
			err:              errProbeReturned,
			expectedFindings: nil,
		},
		{
			name: "no reviews",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{},
							},
							Reviews: []clients.Review{},
							Author:  clients.User{Login: "pedro"},
						},
					},
				},
			},
			expectedFindings: []finding.Finding{
				{
					Probe:   "codeApproved",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "codeReviewed",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "codeReviewTwoReviewers",
					Outcome: finding.OutcomeNegative,
				},
			},
		},
		{
			name: "all authors are bots",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{
									SHA: "sha",
									Committer: clients.User{
										Login: "bot",
										IsBot: true,
									},
									Message: "Title\nPiperOrigin-RevId: 444529962",
								},
							},
							Reviews: []clients.Review{},
							Author: clients.User{
								Login: "bot",
								IsBot: true,
							},
						},
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{
									SHA: "sha2",
									Committer: clients.User{
										Login: "bot",
										IsBot: true,
									},
								},
							},
							Reviews: []clients.Review{},
							Author: clients.User{
								Login: "bot",
								IsBot: true,
							},
						},
					},
				},
			},
			expectedFindings: []finding.Finding{
				{
					Probe:   "codeApproved",
					Outcome: finding.OutcomeNotAvailable,
				},
				{
					Probe:   "codeReviewed",
					Outcome: finding.OutcomeNotAvailable,
				},
				{
					Probe:   "codeReviewTwoReviewers",
					Outcome: finding.OutcomeNotAvailable,
				},
			},
		},
		{
			name: "no approvals, reviewed once",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{
									SHA:       "sha",
									Committer: clients.User{Login: "kratos"},
									Message:   "Title\nPiperOrigin-RevId: 444529962",
								},
							},
							Reviews: []clients.Review{
								{
									Author: &clients.User{Login: "loki"},
								},
							},
							Author: clients.User{Login: "kratos"},
						},
					},
				},
			},
			expectedFindings: []finding.Finding{
				{
					Probe:   "codeApproved",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "codeReviewed",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "codeReviewTwoReviewers",
					Outcome: finding.OutcomeNegative,
				},
			},
		},
		{
			name: "four reviewers, only one unique",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{
									SHA:       "sha",
									Committer: clients.User{Login: "kratos"},
									Message:   "Title\nPiperOrigin-RevId: 444529962",
								},
							},
							Reviews: []clients.Review{
								{
									Author: &clients.User{Login: "loki"},
									State:  "APPROVED",
								},
								{
									Author: &clients.User{Login: "loki"},
									State:  "APPROVED",
								},
								{
									Author: &clients.User{Login: "kratos"},
									State:  "APPROVED",
								},
								{
									Author: &clients.User{Login: "kratos"},
									State:  "APPROVED",
								},
							},
							Author: clients.User{Login: "kratos"},
						},
					},
				},
			},
			expectedFindings: []finding.Finding{
				{
					Probe:   "codeApproved",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "codeReviewed",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "codeReviewTwoReviewers",
					Outcome: finding.OutcomeNegative,
				},
			},
		},
		{
			name: "reviewed and approved twice",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{
									SHA:       "sha",
									Committer: clients.User{Login: "kratos"},
									Message:   "Title\nPiperOrigin-RevId: 444529962",
								},
							},
							Reviews: []clients.Review{
								{
									Author: &clients.User{Login: "loki"},
									State:  "APPROVED",
								},
								{
									Author: &clients.User{Login: "baldur"},
									State:  "APPROVED",
								},
							},
							Author: clients.User{Login: "kratos"},
						},
					},
				},
			},
			expectedFindings: []finding.Finding{
				{
					Probe:   "codeApproved",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "codeReviewed",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "codeReviewTwoReviewers",
					Outcome: finding.OutcomePositive,
				},
			},
		},
	}

	for _, tt := range probeTests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := zrunner.Run(tt.rawResults, probes.CodeReview)
			switch {
			case err != nil && tt.err == nil:
				t.Errorf("Uxpected error %v", err)
			case tt.err != nil && err == nil:
				t.Errorf("Expected error %v, got nil", tt.err)
			case res == nil && err == nil:
				t.Errorf("Probe(s) returned nil for both finding and error")
			default:
				for i := range tt.expectedFindings {
					if tt.expectedFindings[i].Outcome != res[i].Outcome {
						t.Errorf("Code-review probe: %v error: test name: \"%v\", wanted outcome %v, got %v",
							res[i].Probe, tt.name, tt.expectedFindings[i].Outcome, res[i].Outcome)
					}
				}
			}
		})
	}
}
