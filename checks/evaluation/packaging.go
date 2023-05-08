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

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

// Packaging applies the score policy for the Packaging check.
func Packaging(name string, dl checker.DetailLogger, r *checker.PackagingData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	pass := false
	for _, p := range r.Packages {
		if p.Msg != nil {
			// This is a debug message. Let's just replay the message.
			dl.Debug(&checker.LogMessage{
				Text: *p.Msg,
			})
			continue
		}

		// Presence of a single non-debug message means the
		// check passes.
		pass = true

		msg, err := createLogMessage(p)
		if err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
		}
		dl.Info(&msg)
	}

	if pass {
		return checker.CreateMaxScoreResult(name,
			"publishing workflow detected")
	}

	dl.Warn(&checker.LogMessage{
		Text: "no GitHub/GitLab publishing workflow detected",
	})

	return checker.CreateInconclusiveResult(name,
		"no published package detected")
}

func createLogMessage(p checker.Package) (checker.LogMessage, error) {
	var msg checker.LogMessage

	if p.Msg != nil {
		return msg, sce.WithMessage(sce.ErrScorecardInternal, "Msg should be nil")
	}

	if p.File == nil {
		return msg, sce.WithMessage(sce.ErrScorecardInternal, "File field is nil")
	}

	if p.File != nil {
		msg.Path = p.File.Path
		msg.Type = p.File.Type
		msg.Offset = p.File.Offset
	}

	if len(p.Runs) == 0 {
		return msg, sce.WithMessage(sce.ErrScorecardInternal, "no run data")
	}

	msg.Text = fmt.Sprintf("GitHub/GitLab publishing workflow used in run %s", p.Runs[0].URL)

	return msg, nil
}
