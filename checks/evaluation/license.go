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
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

func scoreLicenseCriteria(f *checker.LicenseFile,
	dl checker.DetailLogger,
) int {
	var score int
	msg := checker.LogMessage{
		Path:   "",
		Type:   finding.FileTypeNone,
		Text:   "",
		Offset: 1,
	}
	msg.Path = f.File.Path
	msg.Type = finding.FileTypeSource
	// #1 a license file was found.
	score += 6

	// #2 the licence was found at the top-level or LICENSE/ folder.
	switch f.LicenseInformation.Attribution {
	case checker.LicenseAttributionTypeAPI, checker.LicenseAttributionTypeHeuristics:
		// both repoAPI and scorecard (not using the API) follow checks.md
		// for a file to be found it must have been in the correct location
		// award location points.
		score += 3
		msg.Text = "License file found in expected location"
		dl.Info(&msg)
		// for repo attribution prepare warning if not an recognized license"
		msg.Text = "Any licence detected not an FSF or OSI recognized license"
	case checker.LicenseAttributionTypeOther:
		// TODO ascertain location found
		score += 0
		msg.Text = "License file found in unexpected location"
		dl.Warn(&msg)
		// for non repo attribution not the license detection is not supported
		msg.Text = "Detecting license content not supported"
	default:
	}

	// #3 is the license either an FSF or OSI recognized/approved license
	if f.LicenseInformation.Approved {
		score += 1
		msg.Text = "FSF or OSI recognized license"
		dl.Info(&msg)
	} else {
		// message text for this condition set above
		dl.Warn(&msg)
	}
	return score
}

// License applies the score policy for the License check.
func License(name string, dl checker.DetailLogger,
	r *checker.LicenseData,
) checker.CheckResult {
	var score int
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.LicenseFiles == nil || len(r.LicenseFiles) == 0 {
		return checker.CreateMinScoreResult(name, "license file not detected")
	}

	// TODO: although this a loop, the raw checks will only return one licence file
	// when more than one license file can be aggregated into a composite
	// score, that logic can be comprehended here.
	score = 0
	for idx := range r.LicenseFiles {
		score = scoreLicenseCriteria(&r.LicenseFiles[idx], dl)
	}

	return checker.CreateResultWithScore(name, "license file detected", score)
}
