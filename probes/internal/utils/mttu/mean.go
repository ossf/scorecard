// Copyright 2025 OpenSSF Scorecard Authors
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

package mttu

import (
	"errors"
	"fmt"
	"time"

	"github.com/ossf/scorecard/v5/checker"
)

var (
	epoch                      = time.Unix(0, 0).UTC()
	ErrNoDependencies          = errors.New("no dependencies provided")
	ErrNoDependenciesEvaluated = errors.New("no dependencies could be evaluated (all missing usable data)")
)

// decodeDuration returns the duration represented by t where t == epoch + duration.
// Negative and zero values are clamped to 0.
func decodeDuration(t time.Time) time.Duration {
	if t.IsZero() {
		return 0
	}
	d := t.Sub(epoch)
	if d < 0 {
		return 0
	}
	return d
}

// MeanTimeSinceFirstNewer computes the mean of "time since the oldest newer release"
// across the dependencies contained in raw.MTTUDependenciesResults.Dependencies,
// treating dependencies already on the latest release as contributing 0.
//
// Returns:
//   - mean:      the average duration
//   - evaluated: number of dependencies that contributed to the mean
//   - problems:  any human-readable issues (e.g., missing data)
//   - err:       if no dependencies could be evaluated
func MeanTimeSinceFirstNewer(
	raw *checker.RawResults,
) (mean time.Duration, evaluated int, problems []string, err error) {
	deps := raw.MTTUDependenciesResults.Dependencies
	if len(deps) == 0 {
		return 0, 0, nil, ErrNoDependencies
	}

	var total time.Duration
	var count int

	for _, dep := range deps {
		switch {
		case dep.IsLatest != nil && *dep.IsLatest:
			// Latest → contributes 0 duration.
			count++
		default:
			// Not latest (or unknown). Use encoded duration if present.
			d := decodeDuration(dep.TimeSinceOldestReleast)
			if d == 0 && (dep.IsLatest == nil || (dep.IsLatest != nil && !*dep.IsLatest)) {
				state := "unknown"
				if dep.IsLatest != nil && !*dep.IsLatest {
					state = "not latest"
				}
				problems = append(problems,
					fmt.Sprintf("%s@%s (%s): missing or zero time_since_oldest_releast",
						dep.Name, dep.Version, state))
				continue
			}
			total += d
			count++
		}
	}

	if count == 0 {
		return 0, 0, problems, ErrNoDependenciesEvaluated
	}

	mean = time.Duration(int64(total) / int64(count))
	return mean, count, problems, nil
}
