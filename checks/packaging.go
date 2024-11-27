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

package checks

import (
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/evaluation"
	"github.com/ossf/scorecard/v5/checks/raw/github"
	"github.com/ossf/scorecard/v5/checks/raw/gitlab"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/probes"
	"github.com/ossf/scorecard/v5/probes/zrunner"
)

// CheckPackaging is the registered name for Packaging.
const CheckPackaging = "Packaging"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
	}
	if err := registerCheck(CheckPackaging, Packaging, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

// Packaging runs Packaging check.
func Packaging(c *checker.CheckRequest) checker.CheckResult {
	var rawData, rawDataGithub, rawDataGitlab checker.PackagingData
	var err, errG, errGL error

	switch v := c.RepoClient.(type) {
	case *localdir.LocalDirClient:
		// Performing both packaging checks since we dont know when local
		rawDataGithub, errG = github.Packaging(c)
		rawDataGitlab, errGL = gitlab.Packaging(c)
		// Appending results of checks
		rawData.Packages = append(rawData.Packages, rawDataGithub.Packages...)
		rawData.Packages = append(rawData.Packages, rawDataGitlab.Packages...)
		// checking for errors
		if errG != nil {
			err = errG
		} else if errGL != nil {
			err = errGL
		}
	case *githubrepo.Client:
		rawData, err = github.Packaging(c)
	case *gitlabrepo.Client:
		rawData, err = gitlab.Packaging(c)
	default:
		_ = v
	}

	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckPackaging, e)
	}

	pRawResults := getRawResults(c)
	pRawResults.PackagingResults = rawData

	findings, err := zrunner.Run(pRawResults, probes.Packaging)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckPackaging, e)
	}

	ret := evaluation.Packaging(CheckPackaging, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}
