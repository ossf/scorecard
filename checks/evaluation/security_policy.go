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
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

func scoreSecurityCriteria(contentLen, linkedContentLen, urls, emails, discvuls int) (int, string) {
	score := 0
	reason := ""

	// #1: found one linked (email/http) content: score += 3
	//     rationale: someone to collaborate with or link to
	//     information (strong for community)
	if urls >= 1 || emails >= 1 {
		score += 3
		reason += "linked content, "
	}

	// #2: more than one unique (email/http) linked content found: score += 3
	//     rationale: if more than one link, even stronger for the community
	if (urls + emails) > 1 {
		score += 3
		reason = "multiple " + reason
	}

	// #3: more bytes than the sum of the length of all the linked content found: score += 3
	//     rationale: there appears to be information and context around those links
	//     no credit if there is just a link to a site or an email address (those given above)
	//     the test here is that each piece of linked content will likely contain a space
	//     before and after the content (hence the two multiplier)
	if contentLen > 1 && (contentLen > (linkedContentLen + ((urls + emails) * 2))) {
		score += 3
		reason += "text, "
	}

	// #4: found whole number(s) and or match(es) to "Disclos" and or "Vuln": score += 1
	//     rationale: works towards the intent of the security policy file
	//     regarding whom to contact about vuls and disclosures and timing
	//     e.g., we'll disclose, report a vulnerabily, 30 days, etc.
	//     looking for at least 2 hits
	if discvuls > 1 {
		score += 1
		reason += "vulnerability & disclosure instructions"
	}

	if reason == "" {
		reason = "nothing"
	}
	reason = "security policy contains " + reason

	return score, reason
}

// SecurityPolicy applies the score policy for the Security-Policy check.
func SecurityPolicy(name string, dl checker.DetailLogger, r *checker.SecurityPolicyData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.Files == nil || len(r.Files) == 0 {
		// If the file is null or has zero lengths, directly return as not detected.
		return checker.CreateMinScoreResult(name, "security policy file not detected")
	}

	score := 0
	reason := ""
	var err error
	for _, f := range r.Files {
		msg := checker.LogMessage{
			Path:   f.Path,
			Type:   f.Type,
			Offset: f.Offset,
		}
		if msg.Type == checker.FileTypeURL {
			msg.Text = "security policy detected in org repo"
		} else {
			msg.Text = "security policy detected in current repo"
		}

		var contentLen, linkedContentLen, urls, emails, discvuls int
		fmt.Sscanf(f.Snippet, "%d,%d,%d,%d,%d", &contentLen, &linkedContentLen, &urls, &emails, &discvuls)
		score, reason = scoreSecurityCriteria(contentLen, linkedContentLen, urls, emails, discvuls)

		dl.Info(&msg)
	}

	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	} else {
		return checker.CreateResultWithScore(name, reason, score)
	}
}
