// Copyright 2021 Security Scorecard Authors
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
	"errors"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

var (
	errInvalidArgType   = errors.New("invalid arg type")
	errInvalidArgLength = errors.New("invalid arg length")
)

// License applies the score policy for the License check.
func License(name string, dl checker.DetailLogger,
	r *checker.LicenseData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}
	// We expect a single license.
	if len(r.Files) > 1 {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid number of results: %d",
			len(r.Files)))
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if len(r.Files) == 0 {
		return checker.CreateMinScoreResult(name, "license file not detected")
	}

	dl.Info(&checker.LogMessage{
		Path:   r.Files[0].Path,
		Type:   checker.FileTypeSource,
		Offset: 1,
	})

	return checker.CreateMaxScoreResult(name, "license file detected")
}
