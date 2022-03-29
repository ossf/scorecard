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
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
)

// SignedReleases checks for presence of signed release check.
func SignedReleases(c *checker.CheckRequest) (checker.SignedReleasesData, error) {
	releases, err := c.RepoClient.ListReleases()
	if err != nil {
		return checker.SignedReleasesData{}, fmt.Errorf("%w", err)
	}

	var results checker.SignedReleasesData
	for i, r := range releases {
		results.Releases = append(results.Releases,
			checker.Release{
				Tag: r.TagName,
				URL: checker.File{
					Path: r.URL,
				},
			})

		for _, asset := range r.Assets {
			a := checker.ReleaseAsset{
				URL: checker.File{
					Path: asset.URL,
					Type: checker.FileTypeURL,
				},
				Name: asset.Name,
			}
			results.Releases[i].Assets = append(results.Releases[i].Assets, a)
		}
	}

	// Return raw results.
	return results, nil
}
