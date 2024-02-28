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

package format

import (
	"io"
	"sort"
	"time"

	"github.com/ossf/scorecard/v4/docs/checks"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

const logLevel = log.DefaultLevel

func Normalize(r *pkg.ScorecardResult) {
	if r == nil {
		return
	}

	// these fields will change run-to-run, and aren't indicative of behavior changes.
	r.Repo.CommitSHA = ""
	r.Scorecard = pkg.ScorecardInfo{}
	r.Date = time.Time{}

	sort.Slice(r.Checks, func(i, j int) bool {
		return r.Checks[i].Name < r.Checks[j].Name
	})

	for i := range r.Checks {
		check := &r.Checks[i]
		sort.Slice(check.Details, func(i, j int) bool {
			return pkg.DetailToString(&check.Details[i], logLevel) < pkg.DetailToString(&check.Details[j], logLevel)
		})
	}
}

//nolint:wrapcheck
func JSON(r *pkg.ScorecardResult, w io.Writer) error {
	const details = true
	docs, err := checks.Read()
	if err != nil {
		return err
	}
	Normalize(r)
	return r.AsJSON2(details, logLevel, docs, w)
}
