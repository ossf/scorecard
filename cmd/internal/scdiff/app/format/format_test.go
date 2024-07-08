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

package format

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

func TestJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b scorecard.ScorecardResult
	}{
		{
			name: "repo commit SHA standardized",
			a: scorecard.ScorecardResult{
				Repo: scorecard.RepoInfo{
					Name:      "github.com/foo/bar",
					CommitSHA: "commit a",
				},
			},
			b: scorecard.ScorecardResult{
				Repo: scorecard.RepoInfo{
					Name:      "github.com/foo/bar",
					CommitSHA: "commit b",
				},
			},
		},
		{
			name: "dates standardized",
			a: scorecard.ScorecardResult{
				Date: time.Now(),
			},
			b: scorecard.ScorecardResult{
				Date: time.Now().AddDate(0, 0, -1),
			},
		},
		{
			name: "scorecard info standardized",
			a: scorecard.ScorecardResult{
				Scorecard: scorecard.ScorecardInfo{
					Version:   "version a",
					CommitSHA: "scorecard commit x",
				},
			},
			b: scorecard.ScorecardResult{
				Scorecard: scorecard.ScorecardInfo{
					Version:   "version b",
					CommitSHA: "scorecard commit y",
				},
			},
		},
		{
			name: "check order standardized",
			a: scorecard.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Name:  "Token-Permissions",
						Score: 10,
					},
					{
						Name:  "License",
						Score: 10,
					},
				},
			},
			b: scorecard.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Name:  "License",
						Score: 10,
					},
					{
						Name:  "Token-Permissions",
						Score: 10,
					},
				},
			},
		},
		{
			name: "detail order standardized",
			a: scorecard.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Name:  "Token-Permissions",
						Score: 10,
						Details: []checker.CheckDetail{
							{
								Msg: checker.LogMessage{
									Text: "foo",
								},
								Type: checker.DetailInfo,
							},
							{
								Msg: checker.LogMessage{
									Text: "bar",
								},
								Type: checker.DetailWarn,
							},
						},
					},
				},
			},
			b: scorecard.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Name:  "Token-Permissions",
						Score: 10,
						Details: []checker.CheckDetail{
							{
								Msg: checker.LogMessage{
									Text: "bar",
								},
								Type: checker.DetailWarn,
							},
							{
								Msg: checker.LogMessage{
									Text: "foo",
								},
								Type: checker.DetailInfo,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var bufA, bufB bytes.Buffer
			err := JSON(&tt.a, &bufA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = JSON(&tt.b, &bufB)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if bufA.String() != bufB.String() {
				t.Errorf("outputs not identical: %s", cmp.Diff(bufA.String(), bufB.String()))
			}
		})
	}
}

func Test_normalize_nil_safe(t *testing.T) {
	t.Parallel()
	var x, y *scorecard.ScorecardResult
	Normalize(x)
	Normalize(y)
	if !cmp.Equal(x, y) {
		t.Errorf("normalized results differ: %v", cmp.Diff(x, y))
	}
}
