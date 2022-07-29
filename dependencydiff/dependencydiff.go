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
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
	"github.com/ossf/scorecard/v4/policy"
)

// Depdiff is the exported name for dependency-diff.
const Depdiff = "Dependency-diff"

// A private context struct used for GetDependencyCheckResults.
type dependencydiffContext struct {
	logger                          *sclog.Logger
	ownerName, repoName, base, head string
	ctx                             context.Context
	ghRepo                          clients.Repo
	ghRepoClient                    clients.RepoClient
	ossFuzzClient                   clients.RepoClient
	vulnsClient                     clients.VulnerabilitiesClient
	ciiClient                       clients.CIIBestPracticesClient
	changeTypesToCheck              []string
	checkNamesToRun                 []string
	dependencydiffs                 []dependency
	results                         []pkg.DependencyCheckResult
}

// GetDependencyDiffResults gets dependency changes between two given code commits BASE and HEAD
// along with the Scorecard check results of the dependencies, and returns a slice of DependencyCheckResult.
// TO use this API, an access token must be set. See https://github.com/ossf/scorecard#authentication.
func GetDependencyDiffResults(
	ctx context.Context,
	repoURI string, /* Use the format "ownerName/repoName" as the repo URI, such as "ossf/scorecard". */
	base, head string, /* Two code commits base and head, can use either SHAs or branch names. */
	checksToRun []string, /* A list of enabled check names to run. */
	changeTypes []string, /* A list of dependency change types for which we surface scorecard results. */
) ([]pkg.DependencyCheckResult, error) {

	logger := sclog.NewLogger(sclog.DefaultLevel)
	ownerAndRepo := strings.Split(repoURI, "/")
	if len(ownerAndRepo) != 2 {
		return nil, fmt.Errorf("%w: repo uri input", errInvalid)
	}
	owner, repo := ownerAndRepo[0], ownerAndRepo[1]
	dCtx := dependencydiffContext{
		logger:             logger,
		ownerName:          owner,
		repoName:           repo,
		base:               base,
		head:               head,
		ctx:                ctx,
		changeTypesToCheck: changeTypes,
		checkNamesToRun:    checksToRun,
	}
	// Fetch the raw dependency diffs. This API will also handle error cases such as invalid base or head.
	err := fetchRawDependencyDiffData(&dCtx)
	if err != nil {
		return nil, fmt.Errorf("error in fetchRawDependencyDiffData: %w", err)
	}
	// Map the ecosystem naming convention from GitHub to OSV.
	err = mapDependencyEcosystemNaming(dCtx.dependencydiffs)
	if err != nil {
		return nil, fmt.Errorf("error in mapDependencyEcosystemNaming: %w", err)
	}
	err = getScorecardCheckResults(&dCtx)
	if err != nil {
		return nil, fmt.Errorf("error getting scorecard check results: %w", err)
	}
	return dCtx.results, nil
}

func mapDependencyEcosystemNaming(deps []dependency) error {
	for i := range deps {
		if deps[i].Ecosystem == nil {
			continue
		}
		mappedEcosys, err := toEcosystem(*deps[i].Ecosystem)
		if err != nil {
			wrappedErr := fmt.Errorf("error mapping dependency ecosystem: %w", err)
			return wrappedErr
		}
		deps[i].Ecosystem = asPointer(string(mappedEcosys))

	}
	return nil
}

func initRepoAndClientByChecks(dCtx *dependencydiffContext, dSrcRepo string) error {
	repo, repoClient, ossFuzzClient, ciiClient, vulnsClient, err := checker.GetClients(
		dCtx.ctx, dSrcRepo, "", dCtx.logger,
	)
	if err != nil {
		return fmt.Errorf("error getting the github repo and clients: %w", err)
	}
	dCtx.ghRepo = repo
	dCtx.ghRepoClient = repoClient
	// If the caller doesn't specify the checks to run, run all the checks and return all the clients.
	if dCtx.checkNamesToRun == nil || len(dCtx.checkNamesToRun) == 0 {
		dCtx.ossFuzzClient, dCtx.ciiClient, dCtx.vulnsClient = ossFuzzClient, ciiClient, vulnsClient
		return nil
	}
	for _, cn := range dCtx.checkNamesToRun {
		switch cn {
		case checks.CheckFuzzing:
			dCtx.ossFuzzClient = ossFuzzClient
		case checks.CheckCIIBestPractices:
			dCtx.ciiClient = ciiClient
		case checks.CheckVulnerabilities:
			dCtx.vulnsClient = vulnsClient
		}
	}
	return nil
}

func getScorecardCheckResults(dCtx *dependencydiffContext) error {
	// Initialize the checks to run from the caller's input.
	checksToRun, err := policy.GetEnabled(nil, dCtx.checkNamesToRun, nil)
	if err != nil {
		return fmt.Errorf("error init scorecard checks: %w", err)
	}
	for _, d := range dCtx.dependencydiffs {
		depCheckResult := pkg.DependencyCheckResult{
			PackageURL:       d.PackageURL,
			SourceRepository: d.SourceRepository,
			ChangeType:       d.ChangeType,
			ManifestPath:     d.ManifestPath,
			Ecosystem:        d.Ecosystem,
			Version:          d.Version,
			Name:             d.Name,
			/* The scorecard check result is nil now. */
		}
		if d.ChangeType == nil {
			// Since we allow a dependency having a nil change type, so we also
			// give such a dependency a nil scorecard result.
			dCtx.results = append(dCtx.results, depCheckResult)
			continue
		}
		// (1) If no change types are specified, run the checks on all types of dependencies.
		// (2) If there are change types specified by the user, run the checks on the specified types.
		noneGivenOrIsSpecified := (dCtx.changeTypesToCheck == nil || len(dCtx.changeTypesToCheck) == 0) || /* None specified.*/
			isSpecifiedByUser(*d.ChangeType, dCtx.changeTypesToCheck) /* Specified by the user.*/
		// For now we skip those without source repo urls.
		// TODO (#2063): use the BigQuery dataset to supplement null source repo URLs to fetch the Scorecard results for them.
		if d.SourceRepository != nil && noneGivenOrIsSpecified {
			// Initialize the repo and client(s) corresponding to the checks to run.
			err = initRepoAndClientByChecks(dCtx, *d.SourceRepository)
			if err != nil {
				return fmt.Errorf("error init repo and clients: %w", err)
			}

			// Run scorecard on those types of dependencies that the caller would like to check.
			// If the input map changeTypesToCheck is empty, by default, we run the checks for all valid types.
			// TODO (#2064): use the Scorecare REST API to retrieve the Scorecard result statelessly.
			scorecardResult, err := pkg.RunScorecards(
				dCtx.ctx,
				dCtx.ghRepo,
				// TODO (#2065): In future versions, ideally, this should be
				// the commitSHA corresponding to d.Version instead of HEAD.
				clients.HeadSHA,
				checksToRun,
				dCtx.ghRepoClient,
				dCtx.ossFuzzClient,
				dCtx.ciiClient,
				dCtx.vulnsClient,
			)
			// If the run fails, we leave the current dependency scorecard result empty and record the error
			// rather than letting the entire API return nil since we still expect results for other dependencies.
			if err != nil {
				wrappedErr := sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("scorecard running failed for %s: %v", d.Name, err))
				dCtx.logger.Error(wrappedErr, "")
				depCheckResult.ScorecardResultWithError.Error = wrappedErr
			} else { // Otherwise, we record the scorecard check results for this dependency.
				depCheckResult.ScorecardResultWithError.ScorecardResult = &scorecardResult
			}
		}
		dCtx.results = append(dCtx.results, depCheckResult)
	}
	return nil
}

func asPointer(s string) *string {
	return &s
}

func isSpecifiedByUser(ct pkg.ChangeType, changeTypes []string) bool {
	if changeTypes == nil || len(changeTypes) == 0 {
		return false
	}
	for _, ctByUser := range changeTypes {
		if string(ct) == ctByUser {
			return true
		}
	}
	return false
}
