// Copyright 2022 Security Scorecard Authors
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

package dependencydiff

import (
	"context"

	"github.com/ossf/scorecard/v4/pkg"
)

// GetDependencyDiffResults gets dependency changes between two given code commits BASE and HEAD
// along with the Scorecard check results of the dependencies, and returns a slice of DependencyCheckResult.
// TO use this API, an access token must be set following https://github.com/ossf/scorecard#authentication.
func GetDependencyDiffResults(
	ownerName, repoName, baseSHA, headSHA string, scorecardChecksNames []string,
) ([]pkg.DependencyCheckResult, error) {
	ctx := context.Background()
	// Fetch dependency diffs using the GitHub Dependency Review API.
	return fetchRawDependencyDiffData(ctx, ownerName, repoName, baseSHA, headSHA, scorecardChecksNames)
}
