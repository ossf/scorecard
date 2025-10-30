// Copyright 2021 OpenSSF Scorecard Authors
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

package clients

import "time"

// TrackedIssueLabels is the list of label names that are tracked for maintainer response.
// To add support for new labels, simply add them to this list.
// This is used by both the GitHub client (to fetch label events) and the raw check
// (to process those events).
var TrackedIssueLabels = []string{
	"bug",
	"security",
	"kind/bug",
	"area/security",
	"area/product security",
}

// Issue represents a thread like GitHub issue comment thread.
type Issue struct {
	URI       *string
	CreatedAt *time.Time
	// When the issue was closed (nil if still open).
	ClosedAt          *time.Time
	Author            *User
	AuthorAssociation *RepoAssociation
	Comments          []IssueComment
	// Full label add/remove history (across re-labelings).
	// Handlers should record only the labels they fetch to keep size bounded.
	LabelEvents []LabelEvent
	// ALL label events (including non-tracked labels) for maintainer activity detection.
	AllLabelEvents []LabelEvent
	// State change events (close/reopen) with actor information.
	StateChangeEvents []StateChangeEvent

	// IssueNumber is the platform-agnostic issue identifier used with API calls
	// (GitHub: issue number; GitLab: IID). This field is required by the
	// MaintainersRespondToBugSecurityIssues check and the issues handlers.
	IssueNumber int
}

// IssueComment represents a comment on an issue.
type IssueComment struct {
	CreatedAt         *time.Time
	Author            *User
	AuthorAssociation *RepoAssociation
	URL               string
	IsMaintainer      bool
}

// NEW: LabelEvent records add/remove history for an issue label.
// Used by the MaintainersRespondToBugSecurityIssues check to rebuild
// intervals for "bug" and "security" labels.
type LabelEvent struct {
	CreatedAt    time.Time
	Label        string
	Actor        string
	Added        bool
	IsMaintainer bool
}

// StateChangeEvent records when an issue was closed or reopened.
type StateChangeEvent struct {
	CreatedAt    time.Time
	Actor        string
	IsMaintainer bool
	// Closed is true for close events, false for reopen events
	Closed bool
}
