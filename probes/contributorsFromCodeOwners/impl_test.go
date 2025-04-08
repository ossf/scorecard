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
package contributorsFromCodeOwners

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
			name: "Test that contributors include all code owners",
			raw: &checker.RawResults{
				ContributorsResults: checker.ContributorsData{
					Contributors: []clients.User{
						{
							Login: "Login1",
						},
						{
							Login: "Login2",
						},
						{
							Login: "Login3",
						},
					},
					CodeOwners: []clients.User{
						{
							Login: "Login2",
						},
						{
							Login: "Login3",
						},
						{
							Login: "Login1",
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
		{
			name: "Test that contributors match no code owners",
			raw: &checker.RawResults{
				ContributorsResults: checker.ContributorsData{
					Contributors: []clients.User{
						{
							Login: "Login1",
						},
						{
							Login: "Login2",
						},
						{
							Login: "Login3",
						},
					},
					CodeOwners: []clients.User{
						{
							Login: "Login4",
						},
						{
							Login: "Login5",
						},
						{
							Login: "Login6",
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "Test that contributors match some code owners",
			raw: &checker.RawResults{
				ContributorsResults: checker.ContributorsData{
					Contributors: []clients.User{
						{
							Login: "Login1",
						},
						{
							Login: "Login2",
						},
						{
							Login: "Login3",
						},
					},
					CodeOwners: []clients.User{
						{
							Login: "Login3",
						},
						{
							Login: "Login5",
						},
						{
							Login: "Login6",
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "Test no code owners",
			raw: &checker.RawResults{
				ContributorsResults: checker.ContributorsData{
					Contributors: []clients.User{
						{
							Login: "Login1",
						},
						{
							Login: "Login2",
						},
						{
							Login: "Login3",
						},
					},
					CodeOwners: []clients.User{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "Test no code contributors",
			raw: &checker.RawResults{
				ContributorsResults: checker.ContributorsData{
					CodeOwners: []clients.User{
						{
							Login: "Login3",
						},
						{
							Login: "Login5",
						},
						{
							Login: "Login6",
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
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
				t.Error(err)
			}

			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}
