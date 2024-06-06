// Copyright 2020 OpenSSF Scorecard Authors
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

	"github.com/ossf/scorecard/v5/checker"
)

// SignedReleases checks for presence of signed release check.
func SignedReleases(c *checker.CheckRequest) (checker.SignedReleasesData, error) {
	releases, err := c.RepoClient.ListReleases()
	if err != nil {
		return checker.SignedReleasesData{}, fmt.Errorf("%w", err)
	}

	pkgs := []checker.ProjectPackage{}
	versions, err := c.ProjectClient.GetProjectPackageVersions(c.Ctx, c.Repo.Host(), c.Repo.Path())
	if err != nil {
		c.Dlogger.Debug(&checker.LogMessage{Text: fmt.Sprintf("GetProjectPackageVersions: %v", err)})
		return checker.SignedReleasesData{
			Releases: releases,
			Packages: pkgs,
		}, nil
	}

	for _, v := range versions.Versions {
		prov := checker.PackageProvenance{}

		if len(v.SLSAProvenances) > 0 {
			prov = checker.PackageProvenance{
				Commit:     v.SLSAProvenances[0].Commit,
				IsVerified: v.SLSAProvenances[0].Verified,
			}
		}

		pkgs = append(pkgs, checker.ProjectPackage{
			System:     v.VersionKey.System,
			Name:       v.VersionKey.Name,
			Version:    v.VersionKey.Version,
			Provenance: prov,
		})
	}

	return checker.SignedReleasesData{
		Releases: releases,
		Packages: pkgs,
	}, nil
}
