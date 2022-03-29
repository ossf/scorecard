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

package raw

import (
	"errors"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

var errEmptyClient = errors.New("CII client is nil")

// CIIBestPractices retrieves the raw data for the CIIBestPractices check.
func CIIBestPractices(c *checker.CheckRequest) (checker.CIIBestPracticesData, error) {
	var results checker.CIIBestPracticesData
	if c.CIIClient == nil {
		return results, fmt.Errorf("%w", errEmptyClient)
	}

	badge, err := c.CIIClient.GetBadgeLevel(c.Ctx, c.Repo.URI())
	if err != nil {
		return results, fmt.Errorf("%w", err)
	}

	switch badge {
	case clients.NotFound:
		results.Badge = checker.CIIBadgeNotFound
	case clients.InProgress:
		results.Badge = checker.CIIBadgeInProgress
	case clients.Passing:
		results.Badge = checker.CIIBadgePassing
	case clients.Silver:
		results.Badge = checker.CIIBadgeSilver
	case clients.Gold:
		results.Badge = checker.CIIBadgeGold
	case clients.Unknown:
		results.Badge = checker.CIIBadgeUnknown
	}

	return results, nil
}
