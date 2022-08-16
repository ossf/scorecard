// Copyright 2021 Security Scorecard Authors
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

package evaluation

import (
	"fmt"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	lookBackDays    = 90
	activityPerWeek = 1
	daysInOneWeek   = 7
)

// Maintained applies the score policy for the Maintained check.
func Maintained(name string, dl checker.DetailLogger, r *checker.MaintainedData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if r.ArchivedStatus.Status {
		return checker.CreateMinScoreResult(name, "repo is marked as archived")
	}

	// If not explicitly marked archived, look for activity in past `lookBackDays`.
	threshold := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*lookBackDays /*days*/)
	commitsWithinThreshold := 0
	for i := range r.DefaultBranchCommits {
		if r.DefaultBranchCommits[i].CommittedDate.After(threshold) {
			commitsWithinThreshold++
		}
	}

	// Emit a warning if this repo was created recently
	recencyThreshold := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*lookBackDays /*days*/)
	if r.CreatedAt.After(recencyThreshold) {
		dl.Warn(&checker.LogMessage{Text: fmt.Sprintf("repo was created in the last %d days (Created at: %s), please review its contents carefully", lookBackDays, r.CreatedAt.Format(time.RFC3339))})
		daysSinceRepoCreated := int(time.Since(r.CreatedAt).Hours() / 24)
		return checker.CreateMinScoreResult(name, fmt.Sprintf("repo was created %d days ago, not enough maintenance history", daysSinceRepoCreated))
	}

	issuesUpdatedWithinThreshold := 0
	for i := range r.Issues {
		if hasActivityByCollaboratorOrHigher(&r.Issues[i], threshold) {
			issuesUpdatedWithinThreshold++
		}
	}

	return checker.CreateProportionalScoreResult(name, fmt.Sprintf(
		"%d commit(s) out of %d and %d issue activity out of %d found in the last %d days",
		commitsWithinThreshold, len(r.DefaultBranchCommits), issuesUpdatedWithinThreshold, len(r.Issues), lookBackDays),
		commitsWithinThreshold+issuesUpdatedWithinThreshold, activityPerWeek*lookBackDays/daysInOneWeek)
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
