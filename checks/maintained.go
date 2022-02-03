// Copyright 2020 Security Scorecard Authors
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

package checks

import (
	"fmt"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	// CheckMaintained is the exported check name for Maintained.
	CheckMaintained = "Maintained"
	lookBackDays    = 90
	activityPerWeek = 1
	daysInOneWeek   = 7
)

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckMaintained, IsMaintained); err != nil {
		// this should never happen
		panic(err)
	}
}

// IsMaintained runs Maintained check.
func IsMaintained(c *checker.CheckRequest) checker.CheckResult {
	archived, err := c.RepoClient.IsArchived()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}
	if archived {
		return checker.CreateMinScoreResult(CheckMaintained, "repo is marked as archived")
	}

	// If not explicitly marked archived, look for activity in past `lookBackDays`.
	threshold := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*lookBackDays /*days*/)

	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}
	commitsWithinThreshold := 0
	for i := range commits {
		if commits[i].CommittedDate.After(threshold) {
			commitsWithinThreshold++
		}
	}

	issues, err := c.RepoClient.ListIssues()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}
	issuesUpdatedWithinThreshold := 0
	for i := range issues {
		if hasActivityByCollaboratorOrHigher(&issues[i], threshold) {
			issuesUpdatedWithinThreshold++
		}
	}

	return checker.CreateProportionalScoreResult(CheckMaintained, fmt.Sprintf(
		"%d commit(s) out of %d and %d issue activity out of %d found in the last %d days",
		commitsWithinThreshold, len(commits), issuesUpdatedWithinThreshold, len(issues), lookBackDays),
		commitsWithinThreshold+issuesUpdatedWithinThreshold, activityPerWeek*lookBackDays/daysInOneWeek)
}

// hasActivityByCollaboratorOrHigher returns true if the issue was created or commented on by an
// owner/collaborator/member since the threshold.
func hasActivityByCollaboratorOrHigher(issue *clients.Issue, threshold time.Time) bool {
	if issue == nil {
		return false
	}
	if isCollaboratorOrHigher(issue.AuthorAssociation) && issue.CreatedAt != nil && issue.CreatedAt.After(threshold) {
		// The creator of the issue is a collaborator or higher.
		return true
	}
	for _, comment := range issue.Comments {
		if isCollaboratorOrHigher(comment.AuthorAssociation) && comment.CreatedAt != nil &&
			comment.CreatedAt.After(threshold) {
			// The author of the comment is a collaborator or higher.
			return true
		}
	}
	return false
}

// isCollaboratorOrHigher returns true if the user is a collaborator or higher.
func isCollaboratorOrHigher(repoAssociation *clients.RepoAssociation) bool {
	if repoAssociation == nil {
		return false
	}
	priviledgedRoles := []clients.RepoAssociation{
		clients.RepoAssociationCollaborator,
		clients.RepoAssociationMember,
		clients.RepoAssociationOwner,
	}
	for _, role := range priviledgedRoles {
		if role == *repoAssociation {
			return true
		}
	}
	return false
}
