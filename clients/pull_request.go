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

import (
	"time"
)

// PullRequest struct represents a PR as returned by RepoClient.
//
//nolint:govet
type PullRequest struct {
	Number   int
	MergedAt time.Time
	HeadSHA  string
	Author   User
	Labels   []Label
	Reviews  []Review
	MergedBy User
}

// Label represents a PR label.
type Label struct {
	Name string
}

// Review represents a PR review.
type Review struct {
	Author *User
	State  string
}
