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

	diff = diff[i+4:]
	i = strings.LastIndex(diff, "\"\"\"")
	if i == -1 {
		return diff
	}

	return diff[:i]
}

func GeneratePatch(f checker.File, content []byte) string {
	src := string(content)
	// TODO: call fix method
	dst := src + "\n    # random change for testing patch diff"

	return parseDiff(cmp.Diff(src, dst))
}
