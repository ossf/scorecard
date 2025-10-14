// Copyright 2026 OpenSSF Scorecard Authors
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

package raw

import (
	"fmt"
	"time"

	"github.com/ossf/scorecard/v5/checker"
)

const (
	// InactiveMaintainersLookbackMonths defines the number of months to look back
	// when determining if a maintainer is inactive.
	InactiveMaintainersLookbackMonths = 6
)

// InactiveMaintainers retrieves the raw data for the Inactive-Maintainers check.
// It collects activity information for maintainers over the past 6 months.
func InactiveMaintainers(cr *checker.CheckRequest) (checker.InactiveMaintainersData, error) {
	c := cr.RepoClient

	// Calculate the cutoff time (6 months ago from now)
	cutoff := time.Now().UTC().AddDate(0, -InactiveMaintainersLookbackMonths, 0)

	// Get maintainer activity from the repository client
	activity, err := c.GetMaintainerActivity(cutoff)
	if err != nil {
		return checker.InactiveMaintainersData{}, fmt.Errorf("GetMaintainerActivity: %w", err)
	}

	return checker.InactiveMaintainersData{
		MaintainerActivity: activity,
	}, nil
}
