// Copyright 2022 OpenSSF Scorecard Authors
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

package finding

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/remediation"
)

type Finding struct {
	RuleName    string
	Outcome     checker.Outcome
	Text        *string
	Location    *checker.Path
	Remediation *remediation.Remediation
}

func FindingNew(rule string) (*Finding, error) {
	r, err := rule.RuleNew(rule)
	if err != nil {
		return nil, err
	}
	return &Finding{
		RuleName:    rule,
		Outcome:     checker.OutcomeNegative,
		Remediation: r.Remediation,
		// TODO: remediation and use the branch / etc
	}, nil
}

func (f *Finding) WithText(text string) *Finding {
	f.Text = &text
	return f
}

func (f *Finding) WithLocation(path checker.Path) *Finding {
	f.Location = &path
	return f
}

func (f *Finding) WithPatch(patch string) *Finding {
	f.Remediation.patch = &patch
	return f
}
