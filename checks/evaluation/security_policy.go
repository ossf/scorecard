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
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

func scoreSecurityCriteria(f checker.File, info []checker.SecurityPolicyInformation, dl checker.DetailLogger) int {
	var urls, emails, discvuls, linkedContentLen, score int
	// for Security-Policy, EndOffset denotes the
	// length of the found policy file
	// (i.e., byte offset at end of file)
	contentLen := int(f.EndOffset)

	for _, i := range info {
		valuelen := 0
		counter := 0
		for _, v := range i.InformationValue {
			valuelen += len(v)
			counter += 1
		}

		switch i.InformationType {
		case checker.SecurityPolicyInformationTypeEmail:
			emails = counter
			linkedContentLen += valuelen
		case checker.SecurityPolicyInformationTypeLink:
			urls = counter
			linkedContentLen += valuelen
		case checker.SecurityPolicyInformationTypeText:
			discvuls = counter
		}
	}

	msg := checker.LogMessage{
		Path: f.Path,
		Type: f.Type,
		Text: "",
	}

	// #1: more than one unique (email/http) linked content found: score += 6
	//     rationale: if more than one link, even stronger for the community
	if (urls + emails) > 0 {
		score += 6
		msg.Text = "Found linked content in security policy"
		dl.Info(&msg)
	} else {
		msg.Text = "no email or URL found in security policy"
		dl.Warn(&msg)
	}

	// #2: more bytes than the sum of the length of all the linked content found: score += 3
	//     rationale: there appears to be information and context around those links
	//     no credit if there is just a link to a site or an email address (those given above)
	//     the test here is that each piece of linked content will likely contain a space
	//     before and after the content (hence the two multiplier)
	if contentLen > 1 && (contentLen > (linkedContentLen + ((urls + emails) * 2))) {
		score += 3
		msg.Text = "Found text in security policy"
		dl.Info(&msg)
	} else {
		msg.Text = "No text (beyond any linked content) found in security policy"
		dl.Warn(&msg)
	}

	// #3: found whole number(s) and or match(es) to "Disclos" and or "Vuln": score += 1
	//     rationale: works towards the intent of the security policy file
	//     regarding whom to contact about vuls and disclosures and timing
	//     e.g., we'll disclose, report a vulnerabily, 30 days, etc.
	//     looking for at least 2 hits
	if discvuls > 1 {
		score += 1
		msg.Text = "Found disclosure, vulnerability, and/or timelines in security policy"
		dl.Info(&msg)
	} else {
		msg.Text = "One or no descriptive hints of disclosure, vulnerability, and/or timelines in security policy"
		dl.Warn(&msg)
	}

	return score
}

// SecurityPolicy applies the score policy for the Security-Policy check.
func SecurityPolicy(name string, dl checker.DetailLogger, r *checker.SecurityPolicyData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.File == (checker.File{}) {
		// If the file is unset, directly return as not detected.
		return checker.CreateMinScoreResult(name, "security policy file not detected")
	}

	score := scoreSecurityCriteria(r.File, r.Information, dl)

	msg := checker.LogMessage{
		Path:   r.File.Path,
		Type:   r.File.Type,
		Offset: r.File.Offset,
	}

	if msg.Type == checker.FileTypeURL {
		msg.Text = "security policy detected in org repo"
	} else {
		msg.Text = "security policy detected in current repo"
	}

	dl.Info(&msg)

	return checker.CreateResultWithScore(name, "security policy file detected", score)
}
