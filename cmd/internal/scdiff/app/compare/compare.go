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

package compare

import "github.com/ossf/scorecard/v4/pkg"

// results should be normalized before comparison.
func Results(r1, r2 *pkg.ScorecardResult) bool {
	if r1 == nil && r2 == nil {
		return true
	}

	if (r1 != nil) != (r2 != nil) {
		return false
	}

	// intentionally not comparing CommitSHA
	if r1.Repo.Name != r2.Repo.Name {
		return false
	}

	if !compareChecks(r1, r2) {
		return false
	}

	// not comparing findings, as we're JSON first for now

	return true
}

func compareChecks(r1, r2 *pkg.ScorecardResult) bool {
	if len(r1.Checks) != len(r2.Checks) {
		return false
	}

	for i := 0; i < len(r1.Checks); i++ {
		if r1.Checks[i].Name != r2.Checks[i].Name {
			return false
		}
		if r1.Checks[i].Score != r2.Checks[i].Score {
			return false
		}
		if r1.Checks[i].Reason != r2.Checks[i].Reason {
			return false
		}
		if len(r1.Checks[i].Details) != len(r2.Checks[i].Details) {
			return false
		}
		for j := 0; j < len(r1.Checks[i].Details); j++ {
			if r1.Checks[i].Details[j].Type != r2.Checks[i].Details[j].Type {
				return false
			}
			// TODO compare detail specifics?
		}
	}

	return true
}
