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
	"embed"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/rule"
)

// FileType is the type of a file.
type FileType int

const (
	// FileTypeNone is a default, not defined.
	// FileTypeNone must be `0`.
	FileTypeNone FileType = iota
	// FileTypeSource is for source code files.
	FileTypeSource
	// FileTypeBinary is for binary files.
	FileTypeBinary
	// FileTypeText is for text files.
	FileTypeText
	// FileTypeURL for URLs.
	FileTypeURL
)

type Location struct {
	Type      FileType
	Value     string
	LineStart *uint
	LineEnd   *uint
	Snippet   *string
}

// Outcome is the result of a finding.
type Outcome string

const (
	// OutcomePositive indicates a positive outcome.
	OutcomePositive Outcome = "Positive"
	// OutcomeNegative indicates a negative outcome.
	OutcomeNegative Outcome = "Negative"
)

type Finding struct {
	RuleName    string
	Outcome     Outcome
	Risk        rule.Risk
	Text        string
	Location    *Location
	Remediation *rule.Remediation
}

func FindingNew(loc embed.FS, ruleID string) (*Finding, error) {
	r, err := rule.RuleNew(loc, ruleID)
	if err != nil {
		return nil, err
	}
	f := &Finding{
		RuleName:    ruleID,
		Outcome:     OutcomeNegative,
		Remediation: r.Remediation,
		// TODO: remediation and use the branch / etc
	}
	if r.Remediation != nil {
		f.Risk = r.Risk
	}
	return f, nil
}

func (f *Finding) WithText(text string) *Finding {
	f.Text = text
	return f
}

func (f *Finding) WithLocation(loc Location) *Finding {
	f.Location = &loc
	return f
}

func (f *Finding) WithPatch(patch string) *Finding {
	f.Remediation.Patch = &patch
	return f
}

func (f *Finding) WithRemediationMetadata(values map[string]string) *Finding {
	if f.Remediation != nil {
		// Replace all dynamic values.
		for k, v := range values {
			fmt.Println("befoer:", f.Remediation.Text)
			f.Remediation.Text = strings.Replace(f.Remediation.Text,
				fmt.Sprintf("${{ %s }}", k), v, -1)
			f.Remediation.Markdown = strings.Replace(f.Remediation.Markdown,
				fmt.Sprintf("${{ %s }}", k), v, -1)
		}
	}
	return f
}
