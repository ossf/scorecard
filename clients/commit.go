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
	"io"
	"time"
)

// Commit represents a Git commit.
type Commit struct {
	CommittedDate          time.Time
	Message                string
	SHA                    string
	AssociatedMergeRequest PullRequest
	Committer              User
}

// CommitIterator returns the next commit if present, else returns io.EOF error.
type CommitIterator interface {
	Next() (Commit, error)
}

var _ CommitIterator = &SliceBackedCommitIterator{}

type SliceBackedCommitIterator struct {
	data  []Commit
	index int
}

func (iter *SliceBackedCommitIterator) Next() (Commit, error) {
	defer func() { iter.index++ }()
	if len(iter.data) > iter.index {
		return iter.data[iter.index], nil
	}
	return Commit{}, io.EOF
}

func NewSliceBackedCommitIterator(data []Commit) CommitIterator {
	return &SliceBackedCommitIterator{
		data:  data,
		index: 0,
	}
}
