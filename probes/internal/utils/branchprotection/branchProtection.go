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

	"github.com/ossf/scorecard/v4/finding"
)

var errWrongValue = errors.New("wrong value, should not happen")

func GetTextOutcomeFromBool(b *bool, nilMsg, trueMsg, falseMsg string) (string, finding.Outcome, error) {
	switch {
	case b == nil:
		return nilMsg, finding.OutcomeNotAvailable, nil
	case *b:
		return trueMsg, finding.OutcomePositive, nil
	case !*b:
		return falseMsg, finding.OutcomeNegative, nil
	}
	return "", finding.OutcomeError, errWrongValue
}

func Uint32LargerThan0(u32 *int32, nilMsg, trueMsg, falseMsg string) (string, finding.Outcome, error) {
	switch {
	case u32 == nil:
		return nilMsg, finding.OutcomeNotAvailable, nil
	case *u32 > 0:
		return trueMsg, finding.OutcomePositive, nil
	case *u32 == 0:
		return falseMsg, finding.OutcomeNegative, nil
	}
	return "", finding.OutcomeError, errWrongValue
}
