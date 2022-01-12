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

package checks

import (
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	// CheckCIIBestPractices is the registered name for CIIBestPractices.
	CheckCIIBestPractices = "CII-Best-Practices"
	silverScore           = 7
	passingScore          = 5
	inProgressScore       = 2
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckCIIBestPractices, CIIBestPractices)
}

// CIIBestPractices runs CII-Best-Practices check.
func CIIBestPractices(c *checker.CheckRequest) checker.CheckResult {
	if c.CIIClient == nil {
		return checker.CreateInconclusiveResult(CheckCIIBestPractices, "CII client is nil")
	}

	// TODO: not supported for local clients.
	badgeLevel, err := c.CIIClient.GetBadgeLevel(c.Ctx, c.Repo.URI())
	if err == nil {
		switch badgeLevel {
		case clients.NotFound:
			return checker.CreateMinScoreResult(CheckCIIBestPractices, "no badge detected")
		case clients.InProgress:
			return checker.CreateResultWithScore(CheckCIIBestPractices, "badge detected: in_progress", inProgressScore)
		case clients.Passing:
			return checker.CreateResultWithScore(CheckCIIBestPractices, "badge detected: passing", passingScore)
		case clients.Silver:
			return checker.CreateResultWithScore(CheckCIIBestPractices, "badge detected: silver", silverScore)
		case clients.Gold:
			return checker.CreateMaxScoreResult(CheckCIIBestPractices, "badge detected: gold")
		case clients.Unknown:
			e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("unsupported badge: %v", badgeLevel))
			return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
		}
	}
	e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
}
