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

package depdiff

import (
	"fmt"
)

type DepDiffContext struct {
	OwnerName   string
	RepoName    string
	BaseSHA     string
	HeadSHA     string
	AccessToken string
}

func GetDependencyDiff(ownerName, repoName, baseSHA, headSHA, accessToken string) (string, error) {
	ctx := DepDiffContext{
		OwnerName:   ownerName,
		RepoName:    repoName,
		BaseSHA:     baseSHA,
		HeadSHA:     headSHA,
		AccessToken: accessToken,
	}

	// Fetch dependency diffs using the GitHub Dependency Review API.
	deps, err := raw.FetchDependencyDiffData(ctx)
	if err != nil {
		return "", err
	}
	fmt.Println(deps)

	return "", nil
}
