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
	"errors"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/evaluation"
	"github.com/ossf/scorecard/v4/checks/raw/github"
	"github.com/ossf/scorecard/v4/checks/raw/gitlab"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/gitlabrepo"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckPackaging is the registered name for Packaging.
const CheckPackaging = "Packaging"

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckPackaging, Packaging, nil); err != nil {
		// this should never happen
		panic(err)
	}
}

// Packaging runs Packaging check.
func Packaging(c *checker.CheckRequest) checker.CheckResult {
	var rawData checker.PackagingData
	var err error

	if _, clientType := c.RepoClient.(*githubrepo.Client); clientType {
		rawData, err = github.Packaging(c)
	} else if _, clientType := c.RepoClient.(*gitlabrepo.Client); clientType {
		rawData, err = gitlab.Packaging(c)
	} else {
		err = errors.New("invalid RepoClient")
	}
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckPackaging, e)
	}

	// Set the raw results.
	if c.RawResults != nil {
		c.RawResults.PackagingResults = rawData
	}

	return evaluation.Packaging(CheckPackaging, c.Dlogger, &rawData)
}
