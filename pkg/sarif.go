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

package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v3/checker"
	docs "github.com/ossf/scorecard/v3/docs/checks"
	sce "github.com/ossf/scorecard/v3/errors"
	spol "github.com/ossf/scorecard/v3/policy"
)

type text struct {
	Text string `json:"text,omitempty"`
}

//nolint
type region struct {
	StartLine   *int `json:"startLine,omitempty"`
	EndLine     *int `json:"endLine,omitempty"`
	StartColumn *int `json:"startColumn,omitempty"`
	EndColumn   *int `json:"endColumn,omitempty"`
	CharOffset  *int `json:"charOffset,omitempty"`
	ByteOffset  *int `json:"byteOffset,omitempty"`
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
	// We may populate it later if we can indicate which config file
	// line was violated by the failing check.
	Message *text `json:"message,omitempty"`
}

//nolint
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

//nolint
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

func maxOffset(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func calculateSeverityLevel(risk string) string {
	// nolint:lll
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
	// nolint:lll
	// https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning#reportingdescriptor-object.
	switch risk {
	case "Critical":
		// nolint:goconst
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

	// This should never happen unless
	// the check is buggy. We want to catch it
	// quickly and not fail silently.
	if details.Msg.Offset < 0 {
		panic("invalid offset")
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
	case checker.FileTypeURL:
		line := maxOffset(1, details.Msg.Offset)
		reg = region{
			StartLine: &line,
			Snippet:   snippet,
		}
	case checker.FileTypeNone:
		// Do nothing.
	case checker.FileTypeSource:
		line := maxOffset(1, details.Msg.Offset)
		reg = region{
			StartLine: &line,
			Snippet:   snippet,
		}
	case checker.FileTypeText:
		// Offset of 0 is acceptable here.
		reg = region{
			CharOffset: &details.Msg.Offset,
			Snippet:    snippet,
		}
	case checker.FileTypeBinary:
		// Offset of 0 is acceptable here.
		reg = region{
			// Note: GitHub does not support this.
			ByteOffset: &details.Msg.Offset,
		}
	}
	return reg
}

func shouldAddLocation(detail *checker.CheckDetail, showDetails bool,
	minScore, score int) bool {
	switch {
	default:
		return false
	case detail.Msg.Path == "",
		!showDetails,
		detail.Type != checker.DetailWarn,
		detail.Msg.Type == checker.FileTypeURL:
		return false
	case score == checker.InconclusiveResultScore:
		return true
	case minScore >= score:
		return true
	}
}

func detailsToLocations(details []checker.CheckDetail,
	showDetails bool, minScore, score int) []location {
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

	detaultLine := 1
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
	return tool{
		Driver: driver{
			Name:           strings.Title(name),
			InformationURI: url,
			SemVersion:     version,
			Rules:          nil,
		},
	}
}

func createSARIFRun(uri, toolName, version, commit string, t time.Time,
	category, runName string) run {
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
	category string) *run {
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
	rRem := fmt.Sprintf("**Remediation**:\n%s", generateRemediationMarkdown(remediation))
	return textToMarkdown(fmt.Sprintf("%s\n\n%s\n\n%s", rRem, rSev, rDesc))
}

func createSARIFRule(checkName, checkID, descURL, longDesc, shortDesc, risk string,
	remediation []string,
	tags []string) rule {
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
	return result{
		RuleID: checkID,
		// https://github.com/microsoft/sarif-tutorials/blob/main/docs/2-Basics.md#level
		// Level:     scoreToLevel(minScore, score),
		RuleIndex: pos,
		Message:   text{Text: message},
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

func computeCategory(repos []string) (string, error) {
	// In terms of sets, local < Git-local < GitHub.
	switch {
	default:
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("repo types not supported: %v", repos))
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
	for _, v := range runs {
		res = append(res, *v)
	}
	return res
}

// AsSARIF outputs ScorecardResult in SARIF 2.1.0 format.
func (r *ScorecardResult) AsSARIF(showDetails bool, logLevel zapcore.Level,
	writer io.Writer, checkDocs docs.Doc, policy *spol.ScorecardPolicy,
	policyFile string) error {
	//nolint
	// https://docs.oasis-open.org/sarif/sarif/v2.1.0/cs01/sarif-v2.1.0-cs01.html.
	// We only support GitHub-supported properties:
	// see https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/sarif-support-for-code-scanning#supported-sarif-output-file-properties,
	// https://github.com/microsoft/sarif-tutorials.
	sarif := createSARIFHeader()
	runs := make(map[string]*run)

	// nolint
	for _, check := range r.Checks {
		doc, err := checkDocs.GetCheck(check.Name)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %v: %s", err, check.Name))
		}

		// We need to create a run entry even if the check is disabled or the policy is satisfied.
		// The reason is the following: if a check has findings and is later fixed by a user,
		// the absence of run for the check will indicate that the check was *not* run,
		// so GitHub would keep the findings in the dahsboard. We don't want that.
		category, err := computeCategory(doc.GetSupportedRepoTypes())
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("computeCategory: %v: %s", err, check.Name))
		}
		run := getOrCreateSARIFRun(runs, category, "https://github.com/ossf/scorecard", "scorecard",
			r.Scorecard.Version, r.Scorecard.CommitSHA, r.Date, "supply-chain")

		// Always add rules to indicae which checks were run.
		// We don't have so many rules, so this should not clobber the output too much.
		// See https://github.com/github/codeql-action/issues/810.
		checkID := check.Name
		rule := createSARIFRule(check.Name, checkID,
			doc.GetDocumentationURL(r.Scorecard.CommitSHA),
			doc.GetDescription(), doc.GetShort(), doc.GetRisk(),
			doc.GetRemediation(), doc.GetTags())
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, rule)

		// Check the policy configuration.
		minScore, enabled, err := getCheckPolicyInfo(policy, check.Name)
		if err != nil {
			return err
		}
		if !enabled {
			continue
		}

		// Skip check that do not violate the policy.
		if check.Score >= minScore {
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
		locs := detailsToLocations(check.Details2, showDetails, minScore, check.Score)

		// Add default location if no locations are present.
		// Note: GitHub needs at least one location to show the results.
		// RuleIndex is the position of the corresponding rule in `run.Tool.Driver.Rules`,
		// so it's the last position for us.
		RuleIndex := len(run.Tool.Driver.Rules) - 1
		if len(locs) == 0 {
			locs = addDefaultLocation(locs, policyFile)
			// Use the `reason` as message.
			cr := createSARIFCheckResult(RuleIndex, checkID, check.Reason, &locs[0])
			run.Results = append(run.Results, cr)
		} else {
			for _, loc := range locs {
				// Use the location's message (check's detail's message) as message.
				cr := createSARIFCheckResult(RuleIndex, checkID, loc.Message.Text, &loc)
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
