// Copyright 2021 OpenSSF Scorecard Authors
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

package evaluation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	minContributionsPerUser    = 5
	numberCompaniesForTopScore = 3
)

// Contributors applies the score policy for the Contributors check.
func Contributors(name string, dl checker.DetailLogger,
	r *checker.ContributorsData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	entities := make(map[string]bool)

	for _, user := range r.Users {
		if user.NumContributions < minContributionsPerUser {
			continue
		}

		for _, org := range user.Organizations {
			entities[org.Login] = true
		}

		for _, comp := range user.Companies {
			entities[comp] = true
		}
	}

	names := []string{}
	for c := range entities {
		names = append(names, c)
	}

	sort.Strings(names)

	if len(name) > 0 {
		dl.Info(&checker.LogMessage{
			Text: fmt.Sprintf("contributors work for %v", strings.Join(names, ",")),
		})
	} else {
		dl.Warn(&checker.LogMessage{
			Text: "no contributors have an org or company",
		})
	}

	reason := fmt.Sprintf("%d different organizations found", len(entities))
	return checker.CreateProportionalScoreResult(name, reason, len(entities), numberCompaniesForTopScore)
}
