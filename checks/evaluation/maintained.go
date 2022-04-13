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
		if r.DefaultBranchCommits[i].CommitDate.After(threshold) {
			commitsWithinThreshold++
		}
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
func hasActivityByCollaboratorOrHigher(issue *checker.Issue, threshold time.Time) bool {
	if issue == nil {
		return false
	}

	if isCollaboratorOrHigher(issue.Author) && issue.CreatedAt != nil && issue.CreatedAt.After(threshold) {
		// The creator of the issue is a collaborator or higher.
		return true
	}
	for _, comment := range issue.Comments {
		if isCollaboratorOrHigher(comment.Author) && comment.CreatedAt != nil &&
			comment.CreatedAt.After(threshold) {
			// The author of the comment is a collaborator or higher.
			return true
		}
	}
	return false
}

// isCollaboratorOrHigher returns true if the user is a collaborator or higher.
func isCollaboratorOrHigher(user *checker.User) bool {
	if user == nil || user.RepoAssociation == nil {
		return false
	}

	priviledgedRoles := []checker.RepoAssociation{
		checker.RepoAssociationOwner,
		checker.RepoAssociationCollaborator,
		checker.RepoAssociationContributor,
		checker.RepoAssociationMember,
	}
	for _, role := range priviledgedRoles {
		if role == *user.RepoAssociation {
			return true
		}
	}
	return false
}
