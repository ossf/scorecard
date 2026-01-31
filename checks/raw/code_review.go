// Copyright 2020 OpenSSF Scorecard Authors
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
	"regexp"
	"strconv"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

var (
	rePhabricatorRevID = regexp.MustCompile(`Differential Revision:[^\r\n]*(D\d+)`)
	rePiperRevID       = regexp.MustCompile(`PiperOrigin-RevId:\s*(\d{3,})`)
)

// CodeReview retrieves the raw data for the Code-Review check.
func CodeReview(c clients.RepoClient) (checker.CodeReviewData, error) {
	// Look at the latest commits.
	commits, err := c.ListCommits()
	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	changesets := getChangesets(commits)

	return checker.CodeReviewData{
		DefaultBranchChangesets: changesets,
	}, nil
}

func getGithubRevisionID(c *clients.Commit) string {
	mr := c.AssociatedMergeRequest
	if !c.AssociatedMergeRequest.MergedAt.IsZero() && mr.Number != 0 {
		return strconv.Itoa(mr.Number)
	}
	return ""
}

func getGithubReviews(c *clients.Commit) (reviews []clients.Review) {
	reviews = []clients.Review{}
	reviews = append(reviews, c.AssociatedMergeRequest.Reviews...)

	if !c.AssociatedMergeRequest.MergedAt.IsZero() {
		reviews = append(reviews, clients.Review{Author: &c.AssociatedMergeRequest.MergedBy, State: "APPROVED"})
	}
	return reviews
}

func getProwReviews(c *clients.Commit) (reviews []clients.Review) {
	reviews = []clients.Review{}
	mr := c.AssociatedMergeRequest

	// Count Prow labels as approvals
	// In Prow: lgtm = code review approval, approved = maintainer approval
	hasLGTM := false
	hasApproved := false

	for _, label := range mr.Labels {
		if label.Name == "lgtm" {
			hasLGTM = true
		}
		if label.Name == "approved" {
			hasApproved = true
		}
	}

	// Create synthetic reviews from Prow labels
	// This allows existing review counting logic to work with Prow
	if hasLGTM {
		reviews = append(reviews, clients.Review{
			Author: &clients.User{Login: "prow-lgtm"},
			State:  "APPROVED",
		})
	}
	if hasApproved {
		reviews = append(reviews, clients.Review{
			Author: &clients.User{Login: "prow-approved"},
			State:  "APPROVED",
		})
	}

	return reviews
}

func getGithubAuthor(c *clients.Commit) (author clients.User) {
	return c.AssociatedMergeRequest.Author
}

func getProwRevisionID(c *clients.Commit) string {
	mr := c.AssociatedMergeRequest
	if !c.AssociatedMergeRequest.MergedAt.IsZero() {
		for _, l := range c.AssociatedMergeRequest.Labels {
			if (l.Name == "lgtm" || l.Name == "approved") && mr.Number != 0 {
				return strconv.Itoa(mr.Number)
			}
		}
	}

	return ""
}

func getGerritRevisionID(c *clients.Commit) string {
	m := c.Message
	if strings.Contains(m, "Reviewed-on:") &&
		strings.Contains(m, "Reviewed-by:") {
		return c.SHA
	}
	return ""
}

// Given m, a commit message, find the Phabricator revision ID in it.
func getPhabricatorRevisionID(c *clients.Commit) string {
	m := c.Message

	match := rePhabricatorRevID.FindStringSubmatch(m)
	if len(match) < 2 {
		return ""
	}

	return match[1]
}

// Given m, a commit message, find the piper revision ID in it.
func getPiperRevisionID(c *clients.Commit) string {
	m := c.Message

	match := rePiperRevID.FindStringSubmatch(m)
	if len(match) < 2 {
		return ""
	}

	return match[1]
}

type revisionInfo struct {
	Platform checker.ReviewPlatform
	ID       string
}

func detectCommitRevisionInfo(c *clients.Commit) revisionInfo {
	if revisionID := getProwRevisionID(c); revisionID != "" {
		return revisionInfo{checker.ReviewPlatformProw, revisionID}
	}
	if revisionID := getGithubRevisionID(c); revisionID != "" {
		return revisionInfo{checker.ReviewPlatformGitHub, revisionID}
	}
	if revisionID := getPhabricatorRevisionID(c); revisionID != "" {
		return revisionInfo{checker.ReviewPlatformPhabricator, revisionID}
	}
	if revisionID := getGerritRevisionID(c); revisionID != "" {
		return revisionInfo{checker.ReviewPlatformGerrit, revisionID}
	}
	if revisionID := getPiperRevisionID(c); revisionID != "" {
		return revisionInfo{checker.ReviewPlatformPiper, revisionID}
	}

	return revisionInfo{checker.ReviewPlatformUnknown, ""}
}

// Group commits by the changeset they belong to
// Commits must be in-order.
func getChangesets(commits []clients.Commit) []checker.Changeset {
	changesets := []checker.Changeset{}

	if len(commits) == 0 {
		return changesets
	}

	changesetsByRevInfo := make(map[revisionInfo]checker.Changeset)

	for i := range commits {
		rev := detectCommitRevisionInfo(&commits[i])
		if rev.ID == "" {
			rev.ID = commits[i].SHA
		}

		if changeset, ok := changesetsByRevInfo[rev]; !ok {
			newChangeset := checker.Changeset{
				ReviewPlatform: rev.Platform,
				RevisionID:     rev.ID,
				Commits:        []clients.Commit{commits[i]},
			}

			switch rev.Platform {
			case checker.ReviewPlatformGitHub:
				newChangeset.Reviews = getGithubReviews(&commits[i])
				newChangeset.Author = getGithubAuthor(&commits[i])
			case checker.ReviewPlatformProw:
				newChangeset.Reviews = getProwReviews(&commits[i])
				newChangeset.Author = getGithubAuthor(&commits[i])
			}

			changesetsByRevInfo[rev] = newChangeset
		} else {
			// Part of a previously found changeset.
			changeset.Commits = append(changeset.Commits, commits[i])
			changesetsByRevInfo[rev] = changeset
		}
	}

	// Changesets are returned in map order (i.e. randomized)
	for ri := range changesetsByRevInfo {
		changesets = append(changesets, changesetsByRevInfo[ri])
	}

	return changesets
}
