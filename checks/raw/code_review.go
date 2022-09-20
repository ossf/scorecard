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

package raw

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

// CodeReview retrieves the raw data for the Code-Review check.
func CodeReview(c clients.RepoClient) (checker.CodeReviewData, error) {
	// Look at the latest commits.
	commits, err := c.ListCommits()
	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	changesets := getChangesets(commits)

	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

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

func getProwRevisionID(c *clients.Commit) string {
	mr := c.AssociatedMergeRequest
	if !c.AssociatedMergeRequest.MergedAt.IsZero() {
		for _, l := range c.AssociatedMergeRequest.Labels {
			if l.Name == "lgtm" || l.Name == "approved" && mr.Number != 0 {
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
	p, err := regexp.Compile(`Differential Revision:\s*(\w+)`)
	if err != nil {
		return ""
	}

	match := p.FindStringSubmatch(m)
	if match == nil || len(match) < 2 {
		return ""
	}

	return match[1]
}

// Given m, a commit message, find the piper revision ID in it.
func getPiperRevisionID(c *clients.Commit) string {
	m := c.Message
	matchPiperRevID, err := regexp.Compile(`PiperOrigin-RevId:\s*(\d{3,})`)
	if err != nil {
		return ""
	}

	match := matchPiperRevID.FindStringSubmatch(m)
	if match == nil || len(match) < 2 {
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

	return revisionInfo{}
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

		if changeset, ok := changesetsByRevInfo[rev]; !ok {
			newChangeset := checker.Changeset{
				ReviewPlatform: rev.Platform,
				RevisionID:     rev.ID,
				Commits:        []clients.Commit{commits[i]},
			}

			changesetsByRevInfo[rev] = newChangeset
		} else {
			// Part of a previously found changeset.
			changeset.Commits = append(changeset.Commits, commits[i])
			changesetsByRevInfo[rev] = changeset
		}
	}

	// Changesets are returned in map order (i.e. randomized)
	for ri, cs := range changesetsByRevInfo {
		// Ungroup all commits that don't have revision info
		missing := revisionInfo{}
		if ri == missing {
			for i := range cs.Commits {
				c := cs.Commits[i]
				changesets = append(changesets, checker.Changeset{
					Commits: []clients.Commit{c},
				})
			}
		} else {
			changesets = append(changesets, cs)
		}
	}

	return changesets
}
