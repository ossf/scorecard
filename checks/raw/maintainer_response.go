// Copyright 2025 OpenSSF Scorecard Authors
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
	"sort"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

// TrackedLabels is the list of label names that are tracked for maintainer response.
// To add support for new labels, simply add them to this list.
var TrackedLabels = []string{
	"bug",
	"security",
	"kind/bug",
	"area/security",
	"area/product security",
}

// firstMaintainerCommentOnOrAfter returns the earliest maintainer comment time >= start, or nil.
func firstMaintainerCommentOnOrAfter(comments []clients.IssueComment, start time.Time) *time.Time {
	var best *time.Time
	for i := range comments {
		if !comments[i].IsMaintainer || comments[i].CreatedAt == nil {
			continue
		}
		t := *comments[i].CreatedAt
		if t.Before(start) {
			continue
		}
		if best == nil || t.Before(*best) {
			tt := t
			best = &tt
		}
	}
	return best
}

// MaintainerResponse builds label intervals for tracked issue labels and records whether
// there was ANY reaction (any comment, label removal, or closing) after the label was applied.
// The set of tracked labels is defined in clients.TrackedIssueLabels.
func MaintainerResponse(c *checker.CheckRequest) (checker.IssueResponseData, error) {
	out := &checker.IssueResponseData{}
	rc := c.RepoClient
	issues, err := rc.ListIssuesWithHistory()
	if err != nil {
		return *out, sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	now := time.Now()

	for i := range issues {
		item := processIssue(&issues[i], now)
		out.Items = append(out.Items, item)
	}

	return *out, nil
}

// processIssue converts a single Issue into an IssueResponseLag with computed intervals.
func processIssue(is *clients.Issue, now time.Time) checker.IssueResponseLag {
	item := checker.IssueResponseLag{
		IssueURL:    derefStr(is.URI),
		IssueNumber: is.IssueNumber,
		Open:        is.ClosedAt == nil,
		OpenDays:    daysBetween(derefTime(is.CreatedAt, now), derefTime(is.ClosedAt, now)),
	}

	// Sort label events by time asc to build intervals.
	sort.SliceStable(is.LabelEvents, func(i, j int) bool {
		return is.LabelEvents[i].CreatedAt.Before(is.LabelEvents[j].CreatedAt)
	})

	// Build intervals for each tracked label.
	for _, labelName := range clients.TrackedIssueLabels {
		intervals := buildLabelIntervals(is, labelName, now)
		item.HadLabelIntervals = append(item.HadLabelIntervals, intervals...)
	}

	// Convenience flags:
	item.CurrentlyLabeledBugOrSecurity = false
	for _, labelName := range clients.TrackedIssueLabels {
		if currentlyHasLabel(is.LabelEvents, labelName) {
			item.CurrentlyLabeledBugOrSecurity = true
			break
		}
	}
	item.FirstMaintainerCommentAt = firstMaintainerCommentAfter(
		is.Comments,
		derefTime(is.CreatedAt, now),
	)

	return item
}

// buildLabelIntervals creates intervals for a single label name from label events.
func buildLabelIntervals(is *clients.Issue, labelName string, now time.Time) []checker.LabelInterval {
	var intervals []checker.LabelInterval
	var curStart *time.Time

	for _, ev := range is.LabelEvents {
		if ev.Label != labelName {
			continue
		}

		if ev.Added {
			// Start or restart the interval.
			tcopy := ev.CreatedAt
			curStart = &tcopy
			continue
		}

		// Unlabeled: close interval if we had a start.
		if curStart != nil {
			interval := createInterval(*curStart, ev.CreatedAt, is, labelName, false)
			intervals = append(intervals, interval)
			curStart = nil
		}
	}

	// If label is still present (no unlabeled after last add), close to "now".
	if curStart != nil {
		interval := createInterval(*curStart, now, is, labelName, true)
		intervals = append(intervals, interval)
	}

	return intervals
}

// createInterval creates a LabelInterval with reaction detection.
// isOngoing indicates whether the interval is still active (label still present).
//
// Logic for close/reopen:
// 1. If issue closed by creator or maintainer within 180 days and not reopened: no violation.
// 2. If issue reopened: calculate total open time (excluding closed periods).
// 3. Check if total open time exceeds 180 days without maintainer response.
//
//nolint:gocognit,nestif // Close/reopen logic is inherently complex
func createInterval(start, end time.Time, is *clients.Issue, labelName string, isOngoing bool) checker.LabelInterval {
	// First check for maintainer comment response
	reactionAt := firstMaintainerCommentOnOrAfter(is.Comments, start)

	// Check for maintainer label additions/removals after the interval start
	// ANY label action (not just tracked labels) by a maintainer is considered engagement
	maintainerLabelTime := firstMaintainerLabelAdditionOnOrAfter(is.AllLabelEvents, start, end, isOngoing)
	if maintainerLabelTime != nil && (reactionAt == nil || maintainerLabelTime.Before(*reactionAt)) {
		reactionAt = maintainerLabelTime
	}

	// Now handle close/reopen logic
	// Find all state changes within this label interval
	var stateChanges []clients.StateChangeEvent
	for _, sc := range is.StateChangeEvents {
		if !sc.CreatedAt.Before(start) && (isOngoing || !sc.CreatedAt.After(end)) {
			stateChanges = append(stateChanges, sc)
		}
	}

	// Sort state changes by time
	sort.SliceStable(stateChanges, func(i, j int) bool {
		return stateChanges[i].CreatedAt.Before(stateChanges[j].CreatedAt)
	})

	// Calculate total open time and check close conditions
	totalOpenDays := 0
	currentStart := start

	for i, sc := range stateChanges {
		if sc.Closed {
			// Add the open period before this close
			openPeriodEnd := sc.CreatedAt
			if isOngoing || !openPeriodEnd.After(end) {
				totalOpenDays += daysBetween(currentStart, openPeriodEnd)
			}

			// Check if this is the final close (no reopen after it)
			isFinalClose := true
			for j := i + 1; j < len(stateChanges); j++ {
				if !stateChanges[j].Closed { // Found a reopen
					isFinalClose = false
					break
				}
			}

			// If final close within interval by maintainer or creator, check timing
			if isFinalClose {
				daysBeforeClose := daysBetween(start, sc.CreatedAt)
				if daysBeforeClose <= 180 {
					// Use close as response if earlier than comment response
					if reactionAt == nil || sc.CreatedAt.Before(*reactionAt) {
						closeCopy := sc.CreatedAt
						reactionAt = &closeCopy
					}
				}
			}
		} else {
			// Reopened - start counting open time from here
			currentStart = sc.CreatedAt
		}
	}

	// Add final open period if issue is currently open or ongoing
	if isOngoing || (len(stateChanges) == 0) || (!stateChanges[len(stateChanges)-1].Closed) {
		finalEnd := end
		if isOngoing {
			finalEnd = end
		}
		totalOpenDays += daysBetween(currentStart, finalEnd)
	}

	// Use total open days for duration calculation
	durationDays := totalOpenDays
	if durationDays == 0 {
		// Fallback to simple calculation if no state changes
		durationDays = daysBetween(start, end)
	}

	// Unlabel is itself a reaction (triage) - but only count if within 180 days
	// OR if no close/reopen events exist (backward compatibility)
	if !isOngoing {
		unlabelDays := daysBetween(start, end)
		if len(stateChanges) == 0 || unlabelDays <= 180 {
			// Use end time as response if no earlier response exists
			if reactionAt == nil || end.Before(*reactionAt) {
				endCopy := end
				reactionAt = &endCopy
			}
		}
	}

	return checker.LabelInterval{
		Label:               labelName,
		Start:               start,
		End:                 end,
		MaintainerResponded: reactionAt != nil,
		ResponseAt:          reactionAt,
		DurationDays:        durationDays,
	}
}

func currentlyHasLabel(events []clients.LabelEvent, label string) bool {
	var on bool
	for _, ev := range events {
		if ev.Label != label {
			continue
		}
		if ev.Added {
			on = true
		} else {
			on = false
		}
	}
	return on
}

func firstMaintainerCommentAfter(comments []clients.IssueComment, start time.Time) *time.Time {
	var best *time.Time
	for i := range comments {
		if !comments[i].IsMaintainer || comments[i].CreatedAt == nil {
			continue
		}
		t := *comments[i].CreatedAt
		if t.Before(start) {
			continue
		}
		if best == nil || t.Before(*best) {
			tt := t
			best = &tt
		}
	}
	return best
}

func firstMaintainerLabelAdditionOnOrAfter(
	events []clients.LabelEvent, start, end time.Time, isOngoing bool,
) *time.Time {
	var best *time.Time
	for i := range events {
		// Only consider label actions (additions or removals) by maintainers
		// Note: LabelEvents already contains only tracked labels (filtered by client)
		if !events[i].IsMaintainer {
			continue
		}
		t := events[i].CreatedAt
		// Must be on or after start
		if t.Before(start) {
			continue
		}
		// Must be before or at end (unless ongoing)
		if !isOngoing && t.After(end) {
			continue
		}
		if best == nil || t.Before(*best) {
			tt := t
			best = &tt
		}
	}
	return best
}

func daysBetween(a, b time.Time) int {
	if b.Before(a) {
		return 0
	}
	return int(b.Sub(a).Hours() / 24)
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefTime(p *time.Time, fallback time.Time) time.Time {
	if p == nil {
		return fallback
	}
	return *p
}
