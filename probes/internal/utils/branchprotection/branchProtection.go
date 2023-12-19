// Copyright 2023 OpenSSF Scorecard Authors
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

package branchprotection

import (
	"errors"
	"fmt"

	"github.com/ossf/scorecard/v4/finding"
)

var errWrongValue = errors.New("wrong value, should not happen")

func GetTextOutcomeFromBool(b *bool, rule, branchName string) (string, finding.Outcome, error) {
	switch {
	case b == nil:
		msg := fmt.Sprintf("unable to retrieve whether '%s' is required to merge on branch '%s'", rule, branchName)
		return msg, finding.OutcomeNotAvailable, nil
	case *b:
		msg := fmt.Sprintf("'%s' is required to merge on branch '%s'", rule, branchName)
		return msg, finding.OutcomePositive, nil
	case !*b:
		msg := fmt.Sprintf("'%s' is disable on branch '%s'", rule, branchName)
		return msg, finding.OutcomeNegative, nil
	}
	return "", finding.OutcomeError, errWrongValue
}
