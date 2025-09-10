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

package contributorsFromOrgOrCompany

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
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
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "Test that both User.Companies and User.Organizations are included",
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
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		}, {
			name: "Test multiple users",
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
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		}, {
			name: "Test multiple users where one user has insufficient contributions.",
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
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}
