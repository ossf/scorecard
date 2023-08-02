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

package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/pkg"
)

var errFoo = errors.New("arbitrary error")

type resultCounter struct {
	b     bytes.Buffer
	lines int
}

//nolint:wrapcheck
func (rc *resultCounter) Write(p []byte) (n int, err error) {
	rc.lines += bytes.Count(p, []byte("\n"))
	return rc.b.Write(p)
}

type stubRunner struct{}

func (s stubRunner) Run(repo string) (pkg.ScorecardResult, error) {
	switch repo {
	case "errorRepo":
		return pkg.ScorecardResult{}, errFoo
	case "badCheck":
		return pkg.ScorecardResult{
			Checks: []checker.CheckResult{
				{
					Name:  "not a real check",
					Score: 10,
				},
			},
		}, nil
	default:
		return pkg.ScorecardResult{}, nil
	}
}

func Test_generate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		nResults int
		wantErr  bool
	}{
		{
			name: "iterates over repos",
			input: `repoOne
repoTwo`,
			nResults: 2,
			wantErr:  false,
		},
		{
			name: "repo will error",
			input: `validRepo
errorRepo`,
			wantErr: true,
		},
		{
			name: "output fails due to invalid check",
			input: `validRepo
badCheck`,
			wantErr: true,
		},
	}
	var stubRunner stubRunner
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := strings.NewReader(tt.input)
			var output resultCounter
			err := generate(stubRunner, input, &output)
			if (err != nil) != tt.wantErr {
				t.Errorf("generate returned: %v, wanted err: %t", err, tt.wantErr)
			}
			if !tt.wantErr {
				if output.lines != tt.nResults {
					t.Errorf("got %d results, got %d", output.lines, tt.nResults)
				}
			}
		})
	}
}
