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

// nolint:stylecheck
package contributorsFromOrgOrCompany

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
)

type User struct {
	Login            string
	Companies        []string
	Organizations    []User
	NumContributions int
	ID               int64
	IsBot            bool
}

func Test_Run(t *testing.T) {
	t.Parallel()
	// nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "Should be 'comp1', 'comp2', 'Login'",
			raw: &checker.RawResults{
				ContributorsResults: checker.ContributorsData{
					Users: []clients.User{
						{
							Companies: []string{"comp1", "comp2"},
							Organizations: []clients.User{
								{
									Login:            "Login",
									Companies:        []string{"comp3", "comp4"}, // These should not be included
									NumContributions: 10,
								},
							},
							NumContributions: 10,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
			},
		}, {
			name: "Should be 'comp1', 'comp2', 'comp3', 'comp4', 'Login1', 'Login2'",
			raw: &checker.RawResults{
				ContributorsResults: checker.ContributorsData{
					Users: []clients.User{
						{
							Companies: []string{"comp1", "comp2"},
							Organizations: []clients.User{
								{
									Login:            "Login1",
									NumContributions: 10,
								},
							},
							NumContributions: 10,
						},
						{
							Companies: []string{"comp3", "comp4"},
							Organizations: []clients.User{
								{
									Login:            "Login2",
									NumContributions: 10,
								},
							},
							NumContributions: 10,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
			},
		}, {
			name: "Should be 'comp1', 'comp2', 'Login1' V2. The second user has too few contributions to be considered",
			raw: &checker.RawResults{
				ContributorsResults: checker.ContributorsData{
					Users: []clients.User{
						{
							Companies: []string{"comp1", "comp2"},
							Organizations: []clients.User{
								{
									Login:            "Login1",
									NumContributions: 10,
								},
							},
							NumContributions: 10,
						},
						{
							Companies: []string{"comp3", "comp4"},
							Organizations: []clients.User{
								{
									Login:            "Login2",
									NumContributions: 10,
								},
							},
							NumContributions: 2,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// TODO(#https://github.com/ossf/scorecard/issues/3472) Use common validation function.
			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				t.Fatal(err)
				return
			}

			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(len(tt.outcomes), len(findings)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			for i := range tt.outcomes {
				outcome := &tt.outcomes[i]
				f := &findings[i]
				if diff := cmp.Diff(*outcome, f.Outcome); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
