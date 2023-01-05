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

package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/log"
	spol "github.com/ossf/scorecard/v4/policy"
)

type text struct {
	Text string `json:"text,omitempty"`
}

// nolint
type region struct {
	StartLine   *uint `json:"startLine,omitempty"`
	EndLine     *uint `json:"endLine,omitempty"`
	StartColumn *uint `json:"startColumn,omitempty"`
	EndColumn   *uint `json:"endColumn,omitempty"`
	CharOffset  *uint `json:"charOffset,omitempty"`
	ByteOffset  *uint `json:"byteOffset,omitempty"`
	// Use pointer to omit the entire structure.
	Snippet *text `json:"snippet,omitempty"`
}

type artifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId,omitempty"`
}

type physicalLocation struct {
	Region           region           `json:"region"`
	ArtifactLocation artifactLocation `json:"artifactLocation"`
}

//nolint:govet
type location struct {
	PhysicalLocation physicalLocation `json:"physicalLocation"`
	//nolint
	// This is optional https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning#location-object.
	Message        *text `json:"message,omitempty"`
	HasRemediation bool  `json:"-"`
}

// nolint
type relatedLocation struct {
	ID               int              `json:"id"`
	PhysicalLocation physicalLocation `json:"physicalLocation"`
	Message          text             `json:"message"`
}

type partialFingerprints map[string]string

type defaultConfig struct {
	// "none", "note", "warning", "error",
	// https://github.com/oasis-tcs/sarif-spec/blob/master/Schemata/sarif-schema-2.1.0.json#L1566.
	Level string `json:"level"`
}

type properties struct {
	Precision       string   `json:"precision"`
	ProblemSeverity string   `json:"problem.severity"`
	SeverityLevel   string   `json:"security-severity"`
	Tags            []string `json:"tags"`
}

type help struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown,omitempty"`
}

type rule struct {
	// ID should be an opaque, stable ID.
	// TODO: check if GitHub follows this.
	// Last time I tried, it used ID to display to users.
	ID string `json:"id"`
	// Name must be readable by human.
	Name          string        `json:"name"`
	HelpURI       string        `json:"helpUri"`
	ShortDesc     text          `json:"shortDescription"`
	FullDesc      text          `json:"fullDescription"`
	Help          help          `json:"help"`
	DefaultConfig defaultConfig `json:"defaultConfiguration"`
	Properties    properties    `json:"properties"`
}

type driver struct {
	Name           string `json:"name"`
	InformationURI string `json:"informationUri"`
	SemVersion     string `json:"semanticVersion"`
	Rules          []rule `json:"rules,omitempty"`
}

type tool struct {
	Driver driver `json:"driver"`
}

// nolint
type result struct {
	RuleID           string            `json:"ruleId"`
	Level            string            `json:"level,omitempty"` // Optional.
	RuleIndex        int               `json:"ruleIndex"`
	Message          text              `json:"message"`
	Locations        []location        `json:"locations,omitempty"`
	RelatedLocations []relatedLocation `json:"relatedLocations,omitempty"`
	// Logical location: https://github.com/microsoft/sarif-tutorials/blob/main/docs/2-Basics.md#the-locations-array
	// https://docs.oasis-open.org/sarif/sarif/v2.1.0/cs01/sarif-v2.1.0-cs01.html#_Toc16012457.
	// Not supported by GitHub, but possibly useful.
	PartialFingerprints partialFingerprints `json:"partialFingerprints,omitempty"`
}

type automationDetails struct {
	ID string `json:"id"`
}

type run struct {
	AutomationDetails automationDetails `json:"automationDetails"`
	Tool              tool              `json:"tool"`
	// For generated files during analysis. We leave this blank.
	// See https://github.com/microsoft/sarif-tutorials/blob/main/docs/1-Introduction.md#simple-example.
	Artifacts string `json:"artifacts,omitempty"`
	// This MUST never be omitted or set as `nil`.
	Results []result `json:"results"`
}

type sarif210 struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []run  `json:"runs"`
}

func maxOffset(x, y uint) uint {
	if x < y {
		return y
	}
	return x
}

func calculateSeverityLevel(risk string) string {
	//nolint:lll
	// https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning#reportingdescriptor-object.
	// "over 9.0 is critical, 7.0 to 8.9 is high, 4.0 to 6.9 is medium and 3.9 or less is low".
	switch risk {
	case "Critical":
		return "9.0"
	case "High":
		return "7.0"
	case "Medium":
		return "4.0"
	case "Low":
		return "1.0"
	default:
		panic("invalid risk")
	}
}

func generateProblemSeverity(risk string) string {
	//nolint:lll
	// https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning#reportingdescriptor-object.
	switch risk {
	case "Critical":
		//nolint:goconst
		return "error"
	case "High":
		return "error"
	case "Medium":
		return "warning"
	case "Low":
		return "recommendation"
	default:
		panic("invalid risk")
	}
}

func generateDefaultConfig(risk string) string {
	// "none", "note", "warning", "error",
	return "error"
}

func detailToRegion(details *checker.CheckDetail) region {
	var reg region
	var snippet *text
	if details.Msg.Snippet != "" {
		snippet = &text{Text: details.Msg.Snippet}
	}

	// https://github.com/github/codeql-action/issues/754.
	// "code-scanning currently only supports character offset/length and start/end line/columns offsets".

	// https://docs.oasis-open.org/sarif/sarif/v2.0/csprd02/sarif-v2.0-csprd02.html.
	// "3.30.1 General".
	// Line numbers > 0.
	// byteOffset and charOffset >= 0.
	switch details.Msg.Type {
	default:
		panic("invalid")
	case finding.FileTypeURL:
		line := maxOffset(checker.OffsetDefault, details.Msg.Offset)
		reg = region{
			StartLine: &line,
			Snippet:   snippet,
		}
	case finding.FileTypeNone:
		// Do nothing.
	case finding.FileTypeSource:
		startLine := maxOffset(checker.OffsetDefault, details.Msg.Offset)
		endLine := maxOffset(startLine, details.Msg.EndOffset)
		reg = region{
			StartLine: &startLine,
			EndLine:   &endLine,
			Snippet:   snippet,
		}
	case finding.FileTypeText:
		reg = region{
			CharOffset: &details.Msg.Offset,
			Snippet:    snippet,
		}
	case finding.FileTypeBinary:
		reg = region{
			// Note: GitHub does not support ByteOffset, so we also set
			// StartLine.
			ByteOffset: &details.Msg.Offset,
			StartLine:  &details.Msg.Offset,
		}
	}
	return reg
}

func shouldAddLocation(detail *checker.CheckDetail, showDetails bool,
	minScore, score int,
) bool {
	switch {
	default:
		return false
	case detail.Msg.Path == "",
		!showDetails,
		detail.Type != checker.DetailWarn,
		detail.Msg.Type == finding.FileTypeURL:
		return false
	case score == checker.InconclusiveResultScore:
		return true
	case minScore >= score:
		return true
	}
}

func detailsToLocations(details []checker.CheckDetail,
	showDetails bool, minScore, score int,
) []location {
	locs := []location{}

	//nolint
	// Populate the locations.
	// Note https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/sarif-support-for-code-scanning#result-object
	// "Only the first value of this array is used. All other values are ignored."
	if !showDetails {
		return locs
	}

	for i := range details {
		// Avoid `golang Implicit memory aliasing in for loop` gosec error.
		// https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable.
		d := details[i]
		if !shouldAddLocation(&d, showDetails, minScore, score) {
			continue
		}

		loc := location{
			PhysicalLocation: physicalLocation{
				ArtifactLocation: artifactLocation{
					URI:       d.Msg.Path,
					URIBaseID: "%SRCROOT%",
				},
			},
			Message: &text{Text: d.Msg.Text},
		}

		// Add remediaiton information
		if d.Msg.Remediation != nil {
			loc.Message.Text = fmt.Sprintf("%s\nRemediation tip: %s", loc.Message.Text, d.Msg.Remediation.Markdown)
			loc.HasRemediation = true
		}

		// Set the region depending on the file type.
		loc.PhysicalLocation.Region = detailToRegion(&d)
		locs = append(locs, loc)
	}

	return locs
}

func addDefaultLocation(locs []location, policyFile string) []location {
	// No details or no locations.
	// For GitHub to display results, we need to provide
	// a location anyway, regardless of showDetails.
	if len(locs) != 0 {
		return locs
	}

	detaultLine := checker.OffsetDefault
	loc := location{
		PhysicalLocation: physicalLocation{
			ArtifactLocation: artifactLocation{
				URI:       policyFile,
				URIBaseID: "%SRCROOT%",
			},
			Region: region{
				// TODO: set the line to the check if it's overwritten,
				// or to the global policy.
				StartLine: &detaultLine,
			},
		},
	}
	locs = append(locs, loc)

	return locs
}

func createSARIFHeader() sarif210 {
	return sarif210{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs:    []run{},
	}
}

func createSARIFTool(url, name, version string) tool {
	titleCaser := cases.Title(language.English)

	return tool{
		Driver: driver{
			Name:           titleCaser.String(name),
			InformationURI: url,
			SemVersion:     version,
			Rules:          nil,
		},
	}
}

func createSARIFRun(uri, toolName, version, commit string, t time.Time,
	category, runName string,
) run {
	return run{
		Tool:    createSARIFTool(uri, toolName, version),
		Results: []result{},
		//nolint
		// See https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning#runautomationdetails-object.
		AutomationDetails: automationDetails{
			// Time formatting: https://pkg.go.dev/time#pkg-constants.
			ID: fmt.Sprintf("%s/%s/%s", category, runName, fmt.Sprintf("%s-%s", commit, t.Format(time.RFC822Z))),
		},
	}
}

func getOrCreateSARIFRun(runs map[string]*run, runName string,
	uri, toolName, version, commit string, t time.Time,
	category string,
) *run {
	if prun, exists := runs[runName]; exists {
		return prun
	}
	run := createSARIFRun(uri, toolName, version, commit, t, category, runName)
	runs[runName] = &run
	return &run
}

func generateRemediationMarkdown(remediation []string) string {
	r := ""
	for _, s := range remediation {
		r += fmt.Sprintf("- %s\n", s)
	}
	if r == "" {
		panic("no remediation")
	}
	return r[:len(r)-1]
}

func generateMarkdownText(longDesc, risk string, remediation []string) string {
	rSev := fmt.Sprintf("**Severity**: %s", risk)
	rDesc := fmt.Sprintf("**Details**:\n%s", longDesc)
	rRem := fmt.Sprintf("**Remediation (click \"Show more\" below)**:\n%s", generateRemediationMarkdown(remediation))
	return textToMarkdown(fmt.Sprintf("%s\n\n%s\n\n%s", rRem, rSev, rDesc))
}

func createSARIFRule(checkName, checkID, descURL, longDesc, shortDesc, risk string,
	remediation []string,
	tags []string,
) rule {
	return rule{
		ID:        checkID,
		Name:      checkName,
		ShortDesc: text{Text: checkName},
		FullDesc:  text{Text: shortDesc},
		HelpURI:   descURL,
		Help: help{
			Text:     shortDesc,
			Markdown: generateMarkdownText(longDesc, risk, remediation),
		},
		DefaultConfig: defaultConfig{
			Level: generateDefaultConfig(risk),
		},
		Properties: properties{
			Tags:            tags,
			Precision:       "high",
			ProblemSeverity: generateProblemSeverity(risk),
			SeverityLevel:   calculateSeverityLevel(risk),
		},
	}
}

func createSARIFCheckResult(pos int, checkID, message string, loc *location) result {
	t := fmt.Sprintf("%s\nClick Remediation section below to solve this issue", message)
	if loc.HasRemediation {
		t = fmt.Sprintf("%s\nClick Remediation section below for further remediation help", message)
	}

	return result{
		RuleID: checkID,
		// https://github.com/microsoft/sarif-tutorials/blob/main/docs/2-Basics.md#level
		// Level:     scoreToLevel(minScore, score),
		RuleIndex: pos,
		Message:   text{Text: t},
		Locations: []location{*loc},
	}
}

func getCheckPolicyInfo(policy *spol.ScorecardPolicy, name string) (minScore int, enabled bool, err error) {
	policies := policy.GetPolicies()
	if _, exists := policies[name]; !exists {
		return 0, false,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("missing policy for check: %s", name))
	}
	cp := policies[name]
	if cp.GetMode() == spol.CheckPolicy_DISABLED {
		return 0, false, nil
	}
	return int(cp.GetScore()), true, nil
}

func contains(l []string, elt string) bool {
	for _, v := range l {
		if v == elt {
			return true
		}
	}
	return false
}

func computeCategory(checkName string, repos []string) (string, error) {
	// In terms of sets, local < Git-local < GitHub.
	switch {
	default:
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("repo types not supported: %v", repos))
	case checkName == checks.CheckBranchProtection:
		// This is a special case to be give us more flexibility to move this check around
		// and run it on different GitHub triggers.
		return strings.ToLower(checks.CheckBranchProtection), nil
	case contains(repos, "local"):
		return "local", nil
	// Note: Git-local is not supported by any checks yet.
	case contains(repos, "Git-local"):
		return "local-scm", nil
	case contains(repos, "GitHub"),
		contains(repos, "GitLab"):
		return "online-scm", nil
	}
}

func createSARIFRuns(runs map[string]*run) []run {
	res := []run{}
	// Sort keys.
	keys := make([]string, 0)
	for k := range runs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Iterate over keys.
	for _, k := range keys {
		res = append(res, *runs[k])
	}
	return res
}

func createCheckIdentifiers(name string) (string, string) {
	// Identifier must be in Pascal case.
	// We keep the check name the same as the one used in the documentation.
	n := strings.ReplaceAll(name, "-", "")
	return name, fmt.Sprintf("%sID", n)
}

func filterOutDetailType(details []checker.CheckDetail, t checker.DetailType) []checker.CheckDetail {
	ret := make([]checker.CheckDetail, 0)
	for i := range details {
		d := details[i]
		if d.Type == t {
			continue
		}
		ret = append(ret, d)
	}
	return ret
}

func messageWithScore(msg string, score int) string {
	return fmt.Sprintf("score is %d: %s", score, msg)
}

func createDefaultLocationMessage(check *checker.CheckResult, score int) string {
	details := filterOutDetailType(check.Details, checker.DetailInfo)
	s, b := detailsToString(details, log.WarnLevel)
	if b {
		// Warning: GitHub UX needs a single `\n` to turn it into a `<br>`.
		return messageWithScore(fmt.Sprintf("%s:\n%s", check.Reason, s), score)
	}
	return messageWithScore(check.Reason, score)
}

// AsSARIF outputs ScorecardResult in SARIF 2.1.0 format.
func (r *ScorecardResult) AsSARIF(showDetails bool, logLevel log.Level,
	writer io.Writer, checkDocs docs.Doc, policy *spol.ScorecardPolicy,
) error {
	//nolint
	// https://docs.oasis-open.org/sarif/sarif/v2.1.0/cs01/sarif-v2.1.0-cs01.html.
	// We only support GitHub-supported properties:
	// see https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/sarif-support-for-code-scanning#supported-sarif-output-file-properties,
	// https://github.com/microsoft/sarif-tutorials.
	sarif := createSARIFHeader()
	runs := make(map[string]*run)

	//nolint
	for _, check := range r.Checks {
		doc, err := checkDocs.GetCheck(check.Name)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %v: %s", err, check.Name))
		}

		sarifCheckName, sarifCheckID := createCheckIdentifiers(check.Name)

		// We need to create a run entry even if the check is disabled or the policy is satisfied.
		// The reason is the following: if a check has findings and is later fixed by a user,
		// the absence of run for the check will indicate that the check was *not* run,
		// so GitHub would keep the findings in the dashboard. We don't want that.
		category, err := computeCategory(sarifCheckName, doc.GetSupportedRepoTypes())
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("computeCategory: %v: %s", err, check.Name))
		}
		run := getOrCreateSARIFRun(runs, category, "https://github.com/ossf/scorecard", "scorecard",
			r.Scorecard.Version, r.Scorecard.CommitSHA, r.Date, "supply-chain")

		// Always add rules to indicate which checks were run.
		// We don't have so many rules, so this should not clobber the output too much.
		// See https://github.com/github/codeql-action/issues/810.
		rule := createSARIFRule(sarifCheckName, sarifCheckID,
			doc.GetDocumentationURL(r.Scorecard.CommitSHA),
			doc.GetDescription(), doc.GetShort(), doc.GetRisk(),
			doc.GetRemediation(), doc.GetTags())
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, rule)

		// Check the policy configuration.
		// Here we need to use check.Name instead of sarifCheckName since
		// we need to original check's name.
		minScore, enabled, err := getCheckPolicyInfo(policy, check.Name)
		if err != nil {
			return err
		}
		if !enabled {
			continue
		}

		// Skip check that do not violate the policy.
		if check.Score >= minScore || check.Score == checker.InconclusiveResultScore {
			continue
		}

		// Unclear what to use for PartialFingerprints.
		// GitHub only uses `primaryLocationLineHash`, which is not properly defined
		// and Appendix B of https://docs.oasis-open.org/sarif/sarif/v2.1.0/cs01/sarif-v2.1.0-cs01.html
		// warns about using line number for fingerprints:
		// "suppose the fingerprint were to include the line number where the result was located, and suppose
		// that after the baseline was constructed, a developer inserted additional lines of code above that
		// location. Then in the next run, the result would occur on a different line, the computed fingerprint
		// would change, and the result management system would erroneously report it as a new result."

		// Create locations.
		locs := detailsToLocations(check.Details, showDetails, minScore, check.Score)

		// Add default location if no locations are present.
		// Note: GitHub needs at least one location to show the results.
		// RuleIndex is the position of the corresponding rule in `run.Tool.Driver.Rules`,
		// so it's the last position for us.
		RuleIndex := len(run.Tool.Driver.Rules) - 1
		if len(locs) == 0 {
			// Note: this is not a valid URI but GitHub still accepts it.
			// See https://sarifweb.azurewebsites.net/Validation to test verification.
			locs = addDefaultLocation(locs, "no file associated with this alert")
			msg := createDefaultLocationMessage(&check, check.Score)
			cr := createSARIFCheckResult(RuleIndex, sarifCheckID, msg, &locs[0])
			run.Results = append(run.Results, cr)
		} else {
			for _, loc := range locs {
				// Use the location's message (check's detail's message) as message.
				msg := messageWithScore(loc.Message.Text, check.Score)
				cr := createSARIFCheckResult(RuleIndex, sarifCheckID, msg, &loc)
				run.Results = append(run.Results, cr)
			}
		}
	}

	// Set the sarif's runs.
	sarif.Runs = createSARIFRuns(runs)

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "   ")
	if err := encoder.Encode(sarif); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	return nil
}
