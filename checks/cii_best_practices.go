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
	"github.com/ossf/scorecard/v5/internal/checkdef"
)

// CheckCIIBestPractices is the registered name for CIIBestPractices.
const CheckCIIBestPractices = "CII-Best-Practices"

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckCIIBestPractices, CIIBestPractices, nil); err != nil {
		// this should never happen
		panic(err)
	}
}

// CIIBestPractices will check if the maintainers have a best practice badge.
func CIIBestPractices(c *checker.CheckRequest) checker.CheckResult {
	return checkdef.CIIBestPractices.Run(c)
}
