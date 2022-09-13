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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

var errScmDetection = errors.New("couldn't detect scm platform from commit msg")

// CodeReview retrieves the raw data for the Code-Review check.
func CodeReview(c clients.RepoClient, dl checker.DetailLogger) (checker.CodeReviewData, error) {
	// Look at the latest commits.
	commits, err := c.ListCommits()
	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	changesets := getChangesets(commits, dl)

	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	return checker.CodeReviewData{
		DefaultBranchChangesets: changesets,
	}, nil
}

func getGithubRevisionID(c *clients.Commit, dl checker.DetailLogger) string {
	mr := c.AssociatedMergeRequest
	if !c.AssociatedMergeRequest.MergedAt.IsZero() {
		dl.Info(&checker.LogMessage{
			Text: fmt.Sprintf("commit %s was reviewed on github #%d approved merge request",
				c.SHA, c.AssociatedMergeRequest.Number),
		})
		if mr.Number != 0 {
			return strconv.Itoa(mr.Number)
		}
	}
	return ""
}

func getProwRevisionID(c *clients.Commit, dl checker.DetailLogger) string {
	mr := c.AssociatedMergeRequest
	if !c.AssociatedMergeRequest.MergedAt.IsZero() {
		for _, l := range c.AssociatedMergeRequest.Labels {
			if l.Name == "lgtm" || l.Name == "approved" {
				dl.Info(&checker.LogMessage{
					Text: fmt.Sprintf("commit %s was reviewed on prow #%d approved merge request",
						c.SHA, c.AssociatedMergeRequest.Number),
				})
				if mr.Number != 0 {
					return strconv.Itoa(mr.Number)
				}
			}
		}
	}

	return ""
}

func getGerritRevisionID(c *clients.Commit, dl checker.DetailLogger) string {
	m := c.Message
	if strings.Contains(m, "Reviewed-on:") &&
		strings.Contains(m, "Reviewed-by:") {
		dl.Info(&checker.LogMessage{
			Text: fmt.Sprintf("commit %s was reviewed on gerrit", c.SHA),
		})
		return c.SHA
	}
	return ""
}

// Given m, a commit message, find the Phabricator revision ID in it.
func getPhabricatorRevisionID(c *clients.Commit, dl checker.DetailLogger) string {
	m := c.Message
	p, err := regexp.Compile(`Differential Revision:\s*(\w+)`)
	if err != nil {
		dl.Debug((&checker.LogMessage{Text: "phabricator revisionID regex compile failed"}))
		return ""
	}
	match := p.FindStringSubmatch(m)
	if match == nil || len(match) < 2 {
		return ""
	}

	dl.Info(&checker.LogMessage{
		Text: fmt.Sprintf("commit %s was reviewed on phabricator revision %s", c.SHA, match[1]),
	})

	return match[1]
}

// Given m, a commit message, find the piper revision ID in it.
func getPiperRevisionID(c *clients.Commit, dl checker.DetailLogger) string {
	m := c.Message
	matchPiperRevID, err := regexp.Compile(`PiperOrigin-RevId:\s*(\d{3,})`)
	if err != nil {
		dl.Debug((&checker.LogMessage{Text: "piper regex compile failed"}))
		return ""
	}

	match := matchPiperRevID.FindStringSubmatch(m)

	if match == nil || len(match) < 2 {
		return ""
	}

	dl.Info(&checker.LogMessage{
		Text: fmt.Sprintf("commit %s reviewed through piper revision %s", c.SHA, match[1]),
	})

	return match[1]
}

func getCommitRevisionByPlatform(
	c *clients.Commit,
	dl checker.DetailLogger,
) (string, string, error) {
	if revisionID := getProwRevisionID(c, dl); revisionID != "" {
		return checker.ReviewPlatformProw, revisionID, nil
	}
	if revisionID := getGithubRevisionID(c, dl); revisionID != "" {
		return checker.ReviewPlatformGitHub, revisionID, nil
	}
	if revisionID := getPhabricatorRevisionID(c, dl); revisionID != "" {
		return checker.ReviewPlatformPhabricator, revisionID, nil
	}
	if revisionID := getGerritRevisionID(c, dl); revisionID != "" {
		return checker.ReviewPlatformGerrit, revisionID, nil
	}
	if revisionID := getPiperRevisionID(c, dl); revisionID != "" {
		return checker.ReviewPlatformPiper, revisionID, nil
	}

	return "", "", errScmDetection
}

// Group commits by the changeset they belong to
// Commits must be in-order.
func getChangesets(commits []clients.Commit, dl checker.DetailLogger) []checker.Changeset {
	changesets := []checker.Changeset{}

	if len(commits) == 0 {
		return changesets
	}

	i := 0
	for {
		if i >= len(commits) {
			break
		}
		//nolint:errcheck
		plat, rev, _ := getCommitRevisionByPlatform(&commits[i], dl)
		j := i + 1
		for {
			if j >= len(commits) {
				changesets = append(changesets, checker.Changeset{
					ReviewPlatform: plat,
					RevisionID:     rev,
					Commits:        commits[i:j],
				})
				break
			}

			plat2, rev2, err := getCommitRevisionByPlatform(&commits[j], dl)

			if err != nil || plat2 != plat || rev2 != rev {
				changesets = append(changesets, checker.Changeset{
					ReviewPlatform: plat,
					RevisionID:     rev,
					Commits:        commits[i:j],
				})
				break
			}

			j += 1
		}
		i = j
	}

	// Add data to Changeset raw result (if available)
	// Do this per changeset, instead of per-commit, to save effort
	for i := range changesets {
		augmentChangeset(&changesets[i])
	}

	return changesets
}

func augmentChangeset(changeset *checker.Changeset) {
	if changeset.ReviewPlatform != checker.ReviewPlatformGitHub {
		return
	}

	changeset.Reviews = changeset.Commits[0].AssociatedMergeRequest.Reviews

	// Pull request creator is primary author
	// TODO: Handle case where a user who isn't the PR creator pushes changes
	// to the PR branch without relying on unsigned Git authorship
	changeset.Authors = []clients.User{
		changeset.Commits[0].AssociatedMergeRequest.Author,
	}
}
