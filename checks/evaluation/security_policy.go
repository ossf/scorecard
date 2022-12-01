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
)

func scoreSecurityCriteria(f checker.File,
	info []checker.SecurityPolicyInformation,
	dl checker.DetailLogger,
) int {
	var urls, emails, discvuls, linkedContentLen, score int

	emails = countSecInfo(info, checker.SecurityPolicyInformationTypeEmail, true)
	urls = countSecInfo(info, checker.SecurityPolicyInformationTypeLink, true)
	discvuls = countSecInfo(info, checker.SecurityPolicyInformationTypeText, false)

	for _, i := range findSecInfo(info, checker.SecurityPolicyInformationTypeEmail, true) {
		linkedContentLen += len(i.InformationValue.Match)
	}
	for _, i := range findSecInfo(info, checker.SecurityPolicyInformationTypeLink, true) {
		linkedContentLen += len(i.InformationValue.Match)
	}

	msg := checker.LogMessage{
		Path: f.Path,
		Type: f.Type,
		Text: "",
	}

	// #1: linked content found (email/http): score += 6
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
	if f.FileSize > 1 && (f.FileSize > uint(linkedContentLen+((urls+emails)*2))) {
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

func countSecInfo(secInfo []checker.SecurityPolicyInformation,
	infoType checker.SecurityPolicyInformationType,
	unique bool,
) int {
	keys := make(map[string]bool)
	count := 0
	for _, entry := range secInfo {
		if _, present := keys[entry.InformationValue.Match]; !present && entry.InformationType == infoType {
			keys[entry.InformationValue.Match] = true
			count += 1
		} else if !unique && entry.InformationType == infoType {
			count += 1
		}
	}
	return count
}

func findSecInfo(secInfo []checker.SecurityPolicyInformation,
	infoType checker.SecurityPolicyInformationType,
	unique bool,
) []checker.SecurityPolicyInformation {
	keys := make(map[string]bool)
	var secList []checker.SecurityPolicyInformation
	for _, entry := range secInfo {
		if _, present := keys[entry.InformationValue.Match]; !present && entry.InformationType == infoType {
			keys[entry.InformationValue.Match] = true
			secList = append(secList, entry)
		} else if !unique && entry.InformationType == infoType {
			secList = append(secList, entry)
		}
	}
	return secList
}

// SecurityPolicy applies the score policy for the Security-Policy check.
func SecurityPolicy(name string, dl checker.DetailLogger, r *checker.SecurityPolicyData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if len(r.PolicyFiles) == 0 {
		// If the file is unset, directly return as not detected.
		return checker.CreateMinScoreResult(name, "security policy file not detected")
	}

	// TODO: although this a loop, the raw checks will only return one security policy
	// when more than one security policy file can be aggregated into a composite
	// score, that logic can be comprehended here.
	score := 0
	for _, spd := range r.PolicyFiles {
		score = scoreSecurityCriteria(spd.File,
			spd.Information, dl)

		msg := checker.LogMessage{
			Path: spd.File.Path,
			Type: spd.File.Type,
		}
		if msg.Type == checker.FileTypeURL {
			msg.Text = "security policy detected in org repo"
		} else {
			msg.Text = "security policy detected in current repo"
		}

		dl.Info(&msg)
	}

	return checker.CreateResultWithScore(name, "security policy file detected", score)
}
