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

package githubrepo

import (
	"time"

	"github.com/ossf/scorecard/v4/clients"
)

func copyBoolPtr(src *bool, dest **bool) {
	if src != nil {
		*dest = new(bool)
		**dest = *src
	}
}

func copyStringPtr(src *string, dest **string) {
	if src != nil {
		*dest = new(string)
		**dest = *src
	}
}

func copyInt32Ptr(src *int32, dest **int32) {
	if src != nil {
		*dest = new(int32)
		**dest = *src
	}
}

func copyTimePtr(src *time.Time, dest **time.Time) {
	if src != nil {
		*dest = new(time.Time)
		**dest = *src
	}
}

func copyStringSlice(src []string, dest *[]string) {
	*dest = make([]string, len(src))
	copy(*dest, src)
}

func copyRepoAssociationPtr(src *clients.RepoAssociation, dest **clients.RepoAssociation) {
	if src != nil {
		*dest = new(clients.RepoAssociation)
		**dest = *src
	}
}
