// Copyright Security Scorecard Authors
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

	"github.com/ossf/scorecard/v4/checker"
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

	return result, nil
}
