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

//nolint:stylecheck
package notCreatedRecently

import (
	"embed"
	"fmt"
	"strconv"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/probes"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func init() {
	err := probes.Register(probes.Probe{
		Name:           Probe,
		Implementation: Run,
	})
	if err != nil {
		panic(err)
	}
}

//go:embed *.yml
var fs embed.FS

const (
	Probe = "notCreatedRecently"

	LookbackDayKey = "lookBackDays"
	lookBackDays   = 90
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.MaintainedResults

	recencyThreshold := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*lookBackDays /*days*/)

	var text string
	var outcome finding.Outcome
	if r.CreatedAt.After(recencyThreshold) {
		text = fmt.Sprintf("Repository was created in last %d days.", lookBackDays)
		outcome = finding.OutcomeNegative
	} else {
		text = fmt.Sprintf("Repository was not created in last %d days.", lookBackDays)
		outcome = finding.OutcomePositive
	}
	f, err := finding.NewWith(fs, Probe, text, nil, outcome)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	f = f.WithValue(LookbackDayKey, strconv.Itoa(lookBackDays))
	return []finding.Finding{*f}, Probe, nil
}
