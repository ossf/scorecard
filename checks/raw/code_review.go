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

func isReviewedOnGitHub(c *clients.Commit, dl checker.DetailLogger) (bool, string) {
	mr := c.AssociatedMergeRequest

	return !mr.MergedAt.IsZero(), strconv.Itoa(mr.Number)

}

func isReviewedOnProw(c *clients.Commit, dl checker.DetailLogger) (bool, string) {
	if !c.AssociatedMergeRequest.MergedAt.IsZero() {
		for _, l := range c.AssociatedMergeRequest.Labels {
			if l.Name == "lgtm" || l.Name == "approved" {
				dl.Debug(&checker.LogMessage{
					Text: fmt.Sprintf("commit %s review was through %s #%d approved merge request",
						c.SHA, checker.ReviewPlatformProw, c.AssociatedMergeRequest.Number),
				})
				return true, ""
			}
		}
	}
	return false, ""
}

func isReviewedOnGerrit(c *clients.Commit, dl checker.DetailLogger) (bool, string) {
	m := c.Message
	if strings.Contains(m, "\nReviewed-on: ") &&
		strings.Contains(m, "\nReviewed-by: ") {
		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("commit %s was approved through %s", c.SHA, checker.ReviewPlatformGerrit),
		})
		return true, ""
	}
	return false, ""
}

// Given m, a commit message, find the Phabricator revision ID in it
func getPhabricatorRevId(m string) (string, error) {
	matchPhabricatorRevId, err := regexp.Compile("^Differential Revision:\\s*(\\w+)\\s+")

	if err != nil {
		return "", err
	}

	match := matchPhabricatorRevId.FindStringSubmatch(m)

	if match == nil || len(match) < 2 {
		return "", errors.New("coudn't find phabricator differential revision ID")
	}

	return match[1], nil
}

func isReviewedOnPhabricator(c *clients.Commit, dl checker.DetailLogger) (bool, string) {

	m := c.Message
	if strings.Contains(m, "\nDifferential Revision: ") &&
		strings.Contains(m, "\nReviewed By: ") {
		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf(
				"commit %s was approved through %s",
				c.SHA,
				checker.ReviewPlatformPhabricator,
			),
		})

		revId, err := getPhabricatorRevId(m)

		if err != nil {
			dl.Debug(&checker.LogMessage{
				Text: fmt.Sprintf(
					"couldn't find phab differential revision in commit message for commit=%s",
					c.SHA,
				),
			})
		}

		return true, revId
	}
	return false, ""
}

// Given m, a commit message, find the piper revision ID in it
func getPiperRevId(m string) (string, error) {
	matchPiperRevId, err := regexp.Compile(".PiperOrigin-RevId\\s+:\\s*(\\d{3,})\\s+")

	if err != nil {
		return "", err
	}

	match := matchPiperRevId.FindStringSubmatch(m)

	if match == nil || len(match) < 2 {
		return "", errors.New("coudn't find piper revision ID")
	}

	return match[1], nil
}

func isReviewedOnPiper(c *clients.Commit, dl checker.DetailLogger) (bool, string) {
	m := c.Message
	if strings.Contains(m, "\nPiperOrigin-RevId: ") {
		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("commit %s was approved through %s", c.SHA, checker.ReviewPlatformPiper),
		})

		revId, err := getPiperRevId(m)

		if err != nil {
			dl.Debug(&checker.LogMessage{
				Text: fmt.Sprintf(
					"couldn't find piper revision in commit message for commit=%s",
					c.SHA,
				),
			})
		}

		return true, revId
	}
	return false, ""
}

// Group commits by the changeset they belong to
// Commits must be in-order
func getChangesets(commits []clients.Commit, dl checker.DetailLogger) []checker.Changeset {
	changesets := []checker.Changeset{}

	if len(commits) == 0 {
		return changesets
	}

	currentReviewPlatform, currentRevision, err := getCommitRevisionByPlatform(
		&commits[0],
		dl,
	)

	if err != nil {
		dl.Debug(&checker.LogMessage{Text: err.Error()})
		changesets = append(
			changesets,
			checker.Changeset{
				RevisionID:     currentRevision,
				Commits:        commits[0:1],
				ReviewPlatform: currentReviewPlatform,
			},
		)
	}

	j := 0
	for i := 0; i < len(commits); i++ {
		if i == len(commits)-1 {
			changesets = append(
				changesets,
				checker.Changeset{
					Commits:        commits[j:i],
					ReviewPlatform: currentReviewPlatform,
					RevisionID:     currentRevision,
				},
			)
			break
		}

		nextReviewPlatform, nextRevision, err := getCommitRevisionByPlatform(&commits[i+1], dl)
		if err != nil || nextReviewPlatform != currentReviewPlatform ||
			nextRevision != currentRevision {
			if err != nil {
				dl.Debug(&checker.LogMessage{Text: err.Error()})
			}
			// Add all previous commits to the 'batch' of a single changeset
			changesets = append(
				changesets,
				checker.Changeset{
					Commits:        commits[j:i],
					ReviewPlatform: currentReviewPlatform,
					RevisionID:     currentRevision,
				},
			)
			currentReviewPlatform = nextReviewPlatform
			currentRevision = nextRevision
			j = i + 1
		}
	}

	// Add data to Changeset raw result (if available)
	// Do this per changeset, instead of per-commit, to save effort
	for _, changeset := range changesets {
		augmentChangeset(&changeset)
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

func getCommitRevisionByPlatform(
	c *clients.Commit,
	dl checker.DetailLogger,
) (string, string, error) {
	foundRev, revisionId := isReviewedOnGitHub(c, dl)
	if foundRev {
		return checker.ReviewPlatformGitHub, revisionId, nil
	}

	foundRev, revisionId = isReviewedOnProw(c, dl)
	if foundRev {
		return checker.ReviewPlatformProw, revisionId, nil
	}

	foundRev, revisionId = isReviewedOnGerrit(c, dl)
	if foundRev {
		return checker.ReviewPlatformGerrit, revisionId, nil
	}

	foundRev, revisionId = isReviewedOnPhabricator(c, dl)
	if foundRev {
		return checker.ReviewPlatformPhabricator, revisionId, nil
	}

	foundRev, revisionId = isReviewedOnPiper(c, dl)
	if foundRev {
		return checker.ReviewPlatformPiper, revisionId, nil
	}

	return "", "", errors.New(
		fmt.Sprintf("couldn't find linked review platform for commit %s", c.SHA),
	)
}
