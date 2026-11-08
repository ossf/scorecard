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

// Package checks defines all Scorecard checks.
package checks

import (
	"os"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/internal/checkdef"
)

// registered pairs each check's definition with the exported function
// that runs it.
var registered = []struct {
	def *checkdef.Definition
	fn  checker.CheckFn
}{
	{checkdef.BinaryArtifacts, BinaryArtifacts},
	{checkdef.BranchProtection, BranchProtection},
	{checkdef.CIIBestPractices, CIIBestPractices},
	{checkdef.CITests, CITests},
	{checkdef.CodeReview, CodeReview},
	{checkdef.Contributors, Contributors},
	{checkdef.DangerousWorkflow, DangerousWorkflow},
	{checkdef.DependencyUpdateTool, DependencyUpdateTool},
	{checkdef.Fuzzing, Fuzzing},
	{checkdef.License, License},
	{checkdef.Maintained, Maintained},
	{checkdef.Packaging, Packaging},
	{checkdef.PinnedDependencies, PinningDependencies},
	{checkdef.SAST, SAST},
	{checkdef.SBOM, SBOM},
	{checkdef.SecurityPolicy, SecurityPolicy},
	{checkdef.SignedReleases, SignedReleases},
	{checkdef.TokenPermissions, TokenPermissions},
	{checkdef.Vulnerabilities, Vulnerabilities},
	{checkdef.Webhooks, WebHooks},
}

// allChecks is the list of all registered security checks.
var allChecks = func() checker.CheckNameToFnMap {
	m := checker.CheckNameToFnMap{}
	for _, r := range registered {
		m[r.def.Name()] = checker.Check{
			Fn:                    r.fn,
			SupportedRequestTypes: r.def.RequestTypes(),
		}
	}
	return m
}()

func getAll(overrideExperimental bool) checker.CheckNameToFnMap {
	// need to make a copy or caller could mutate original map
	possibleChecks := checker.CheckNameToFnMap{}
	for k, v := range allChecks {
		possibleChecks[k] = v
	}

	if overrideExperimental {
		return possibleChecks
	}

	if _, experimental := os.LookupEnv("SCORECARD_EXPERIMENTAL"); !experimental {
		// TODO: remove this check when v6 is released
		for _, r := range registered {
			if r.def.Experimental() {
				delete(possibleChecks, r.def.Name())
			}
		}
	}

	return possibleChecks
}

// GetAll returns the full list of default checks, excluding any experimental checks
// unless environment variable constraints are satisfied.
func GetAll() checker.CheckNameToFnMap {
	return getAll(false /*overrideExperimental*/)
}

// GetAllWithExperimental returns the full list of checks, including experimental checks.
func GetAllWithExperimental() checker.CheckNameToFnMap {
	return getAll(true /*overrideExperimental*/)
}
