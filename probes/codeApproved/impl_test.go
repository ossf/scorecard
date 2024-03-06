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
package codeApproved

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/test"
)

func TestProbeCodeApproved(t *testing.T) {
	t.Parallel()
	probeTests := []struct {
		name             string
		rawResults       *checker.RawResults
		err              error
		expectedOutcomes []finding.Outcome
	}{
		{
			name: "no changesets",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{},
				},
			},
			expectedOutcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "no changesets no authors",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{},
							},
							Reviews: []clients.Review{},
							Author:  clients.User{Login: ""},
						},
					},
				},
			},
			expectedOutcomes: []finding.Outcome{
				finding.OutcomeError,
			},
		},
		{
			name: "no changesets authors bot",
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
							Author: clients.User{
								Login: "bot",
								IsBot: true,
							},
						},
					},
				},
			},
			expectedOutcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "no changesets no review authors",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{},
							},
							Reviews: []clients.Review{
								{
									Author: &clients.User{Login: ""},
								},
							},
							Author: clients.User{Login: "pedro"},
						},
					},
				},
			},
			expectedOutcomes: []finding.Outcome{
				finding.OutcomeError,
			},
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
			expectedOutcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "only approved bot PRs gives not applicable outcome",
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
							Reviews: []clients.Review{
								{
									Author: &clients.User{Login: "baldur"},
									State:  "APPROVED",
								},
							},
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
							Reviews: []clients.Review{
								{
									Author: &clients.User{Login: "baldur"},
									State:  "APPROVED",
								},
							},
							Author: clients.User{
								Login: "bot",
								IsBot: true,
							},
						},
					},
				},
			},
			expectedOutcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
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
			expectedOutcomes: []finding.Outcome{
				finding.OutcomeNegative,
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
			expectedOutcomes: []finding.Outcome{
				finding.OutcomePositive,
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
			expectedOutcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "only unreviewed bot changesets gives negative outcome",
			rawResults: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							ReviewPlatform: checker.ReviewPlatformGitHub,
							Commits: []clients.Commit{
								{
									SHA:       "sha",
									Committer: clients.User{Login: "dependabot"},
									Message:   "foo",
								},
							},
							Reviews: []clients.Review{},
							Author: clients.User{
								IsBot: true,
								Login: "dependabot",
							},
						},
					},
				},
			},
			expectedOutcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
	}

	for _, tt := range probeTests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, probeID, err := Run(tt.rawResults)
			switch {
			case err != nil && tt.err == nil:
				t.Errorf("Uxpected error %v", err)
			case tt.err != nil && err == nil:
				t.Errorf("Expected error %v, got nil", tt.err)
			case res == nil && err == nil:
				t.Errorf("Probe returned nil for both finding and error")
			case probeID != Probe:
				t.Errorf("Probe returned the wrong probe ID")
			default:
				test.AssertOutcomes(t, res, tt.expectedOutcomes)
			}
		})
	}
}
