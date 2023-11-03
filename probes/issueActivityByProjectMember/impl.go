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
package issueActivityByProjectMember

import (
	"embed"
	"fmt"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const (
	lookBackDays = 90
)

const Probe = "issueActivityByProjectMember"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.MaintainedResults
	numberOfIssuesUpdatedWithinThreshold := 0

	// Look for activity in past `lookBackDays`.
	threshold := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*lookBackDays /*days*/)
	var findings []finding.Finding
	for i := range r.Issues {
		if hasActivityByCollaboratorOrHigher(&r.Issues[i], threshold) {
			numberOfIssuesUpdatedWithinThreshold++
		}
	}

	if numberOfIssuesUpdatedWithinThreshold > 0 {
		f, err := finding.NewWith(fs, Probe,
			"Found a issue within the threshold.", nil,
			finding.OutcomePositive)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValues(map[string]int{
			"numberOfIssuesUpdatedWithinThreshold": numberOfIssuesUpdatedWithinThreshold,
		})
		findings = append(findings, *f)
	} else {
		f, err := finding.NewWith(fs, Probe,
			"Did not find issues within the threshold.", nil,
			finding.OutcomeNegative)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}

// hasActivityByCollaboratorOrHigher returns true if the issue was created or commented on by an
// owner/collaborator/member since the threshold.
func hasActivityByCollaboratorOrHigher(issue *clients.Issue, threshold time.Time) bool {
	if issue == nil {
		return false
	}

	if issue.AuthorAssociation.Gte(clients.RepoAssociationCollaborator) &&
		issue.CreatedAt != nil && issue.CreatedAt.After(threshold) {
		// The creator of the issue is a collaborator or higher.
		return true
	}
	for _, comment := range issue.Comments {
		if comment.AuthorAssociation.Gte(clients.RepoAssociationCollaborator) &&
			comment.CreatedAt != nil &&
			comment.CreatedAt.After(threshold) {
			// The author of the comment is a collaborator or higher.
			return true
		}
	}
	return false
}
