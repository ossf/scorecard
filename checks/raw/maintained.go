// Copyright 2022 OpenSSF Scorecard Authors
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

// Maintained checks for maintenance.
func Maintained(c *checker.CheckRequest) (checker.MaintainedData, error) {
	var result checker.MaintainedData

	// Archived status.
	archived, err := c.RepoClient.IsArchived()
	if err != nil {
		return result, fmt.Errorf("%w", err)
	}
	result.ArchivedStatus.Status = archived

	// Recent commits.
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		return result, fmt.Errorf("%w", err)
	}
	result.DefaultBranchCommits = commits

	// Recent issues.
	issues, err := c.RepoClient.ListIssues()
	if err != nil {
		return result, fmt.Errorf("%w", err)
	}
	result.Issues = issues

	createdAt, err := c.RepoClient.GetCreatedAt()
	if err != nil {
		return result, fmt.Errorf("%w", err)
	}
	result.CreatedAt = createdAt

	// Maintainer activity (6-month lookback).
	// This is best-effort; platforms that don't support it will return empty map.
	cutoff := time.Now().UTC().AddDate(0, -6, 0)
	maintainerActivity, err := c.RepoClient.GetMaintainerActivity(cutoff)
	if err != nil {
		// Don't fail the entire check if maintainer activity cannot be retrieved.
		// Some platforms may not support this feature.
		maintainerActivity = make(map[string]bool)
	}
	result.MaintainerActivity = maintainerActivity

	return result, nil
}
