// Copyright 2024 OpenSSF Scorecard Authors
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

package patch

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/ossf/scorecard/v5/checker"
)

func parseDiff(diff string) string {
	i := strings.Index(diff, "\"\"\"\n")
	if i == -1 {
		return diff
	}
	//remove everything before """\n
	diff = diff[i+4:]
	i = strings.LastIndex(diff, "\"\"\"")
	if i == -1 {
		return diff
	}
	//remove everything after \n  \t"""
	return diff[:i]
}

// TODO: Receive the dangerous workflow as parameter
func GeneratePatch(f checker.File) string {
	// TODO: Implement
	// example:
	// type scriptInjection
	// path {.github/workflows/active-elastic-job~active-elastic-job~build.yml  github.head_ref  91 0 0 1}
	// snippet github.head_ref
	src := `asasas
hello """ola"""
	message=$(echo "${{ github.event.head_commit.message }}" | tail -n +3)
adios`
	dst := `asasas
hello """ola"""
	message=$(echo $COMMIT | tail -n +3)
adios`
	return parseDiff(cmp.Diff(src, dst))
}
