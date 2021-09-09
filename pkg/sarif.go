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

	//nolint:gosec
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v2/checker"
	docs "github.com/ossf/scorecard/v2/docs/checks"
	sce "github.com/ossf/scorecard/v2/errors"
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

//TODO: remove linter and update unit tests.
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
	Precision string   `json:"precision"`
	Tags      []string `json:"tags"`
}

type help struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown,omitempty"`
}

type rule struct {
	ID            string        `json:"id"`
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
	Rules          []rule `json:"rules"`
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
	Artifacts string   `json:"artifacts,omitempty"`
	Results   []result `json:"results"`
}

type sarif210 struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []run  `json:"runs"`
}

func detailToRegion(details *checker.CheckDetail) region {
	var reg region
	var snippet *text
	if details.Msg.Snippet != "" {
		snippet = &text{Text: details.Msg.Snippet}
	}
	switch details.Msg.Type {
	default:
		panic("invalid")
	case checker.FileTypeURL:
		reg = region{
			StartLine: &details.Msg.Offset,
			Snippet:   snippet,
		}
	case checker.FileTypeNone:
		// Do nothing.
	case checker.FileTypeSource:
		reg = region{
			StartLine: &details.Msg.Offset,
			Snippet:   snippet,
		}
	case checker.FileTypeText:
		reg = region{
			CharOffset: &details.Msg.Offset,
			Snippet:    snippet,
		}
	case checker.FileTypeBinary:
		reg = region{
			ByteOffset: &details.Msg.Offset,
		}
	}
	return reg
}

func shouldAddLocation(detail *checker.CheckDetail, showDetails bool,
	logLevel zapcore.Level, minScore, score int) bool {
	switch {
	default:
		return false
	case detail.Msg.Path == "",
		!showDetails,
		logLevel != zapcore.DebugLevel && detail.Type == checker.DetailDebug,
		detail.Msg.Type == checker.FileTypeURL:
		return false
	case score == checker.InconclusiveResultScore:
		return true
	case minScore < score && detail.Type == checker.DetailInfo:
		return false
	case minScore >= score:
		return true
	}
}

func shouldAddRelatedLocation(detail *checker.CheckDetail, showDetails bool,
	logLevel zapcore.Level, minScore, score int) bool {
	switch {
	default:
		return false
	case detail.Msg.Path == "",
		!showDetails,
		detail.Msg.Type != checker.FileTypeURL,
		logLevel != zapcore.DebugLevel && detail.Type == checker.DetailDebug:
		return false
	case score == checker.InconclusiveResultScore:
		return true
	case minScore < score && detail.Type == checker.DetailInfo:
		return false
	case minScore >= score:
		return true
	}
}

func detailsToLocations(details []checker.CheckDetail,
	showDetails bool, logLevel zapcore.Level, minScore, score int) []location {
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
		if !shouldAddLocation(&d, showDetails, logLevel, minScore, score) {
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

func addDefaultLocation(locs []location) []location {
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
				URI:       ".github/scorecard.yml",
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

func detailsToRelatedLocations(details []checker.CheckDetail,
	showDetails bool, logLevel zapcore.Level, minScore, score int) []relatedLocation {
	rlocs := []relatedLocation{}

	//nolint
	// Populate the related locations.
	for i, d := range details {
		if !shouldAddRelatedLocation(&d, showDetails, logLevel, minScore, score) {
			continue
		}
		// TODO: logical locations.
		rloc := relatedLocation{
			ID:      i,
			Message: text{Text: d.Msg.Text},
			PhysicalLocation: physicalLocation{
				ArtifactLocation: artifactLocation{
					// Note: We don't set
					// URIBaseID: "PROJECTROOT" because
					// the URL is an `absolute` path.
					URI: d.Msg.Path,
				},
			},
		}
		// Set the region depending on file type.
		rloc.PhysicalLocation.Region = detailToRegion(&d)
		rlocs = append(rlocs, rloc)
	}
	return rlocs
}

// TODO: update using config file.
func scoreToLevel(minScore, score int) string {
	// Any of error, note, warning, none.
	switch {
	default:
		return "error"
	case score == checker.MaxResultScore, score == checker.InconclusiveResultScore,
		minScore <= score:
		return "note"
	}
}

func createSARIFHeader(url, category, name, version, commit string, t time.Time) sarif210 {
	return sarif210{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []run{
			{
				Tool: tool{
					Driver: driver{
						Name:           strings.Title(name),
						InformationURI: url,
						SemVersion:     version,
						Rules:          nil,
					},
				},
				//nolint
				// See https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning#runautomationdetails-object.
				AutomationDetails: automationDetails{
					// Time formatting: https://pkg.go.dev/time#pkg-constants.
					ID: fmt.Sprintf("%s/%s/%s", category, name, fmt.Sprintf("%s-%s", commit, t.Format(time.RFC822Z))),
				},
			},
		},
	}
}

func createSARIFRule(checkName, checkID, descURL, longDesc, shortDesc string,
	tags []string) rule {
	return rule{
		ID:   checkID,
		Name: checkName,
		// TODO: verify this works on GitHub.
		ShortDesc: text{Text: textToHTML(shortDesc)},
		FullDesc:  text{Text: textToHTML(longDesc)},
		HelpURI:   descURL,
		Help: help{
			Text:     longDesc,
			Markdown: textToMarkdown(longDesc),
		},
		DefaultConfig: defaultConfig{
			Level: "error",
		},
		Properties: properties{
			Tags:      tags,
			Precision: "high",
		},
	}
}

func createSARIFResult(pos int, checkID, reason string, minScore, score int,
	locs []location, rlocs []relatedLocation) result {
	return result{
		RuleID: checkID,
		// https://github.com/microsoft/sarif-tutorials/blob/main/docs/2-Basics.md#level
		Level:            scoreToLevel(minScore, score),
		RuleIndex:        pos,
		Message:          text{Text: reason},
		Locations:        locs,
		RelatedLocations: rlocs,
	}
}

// AsSARIF outputs ScorecardResult in SARIF 2.1.0 format.
func (r *ScorecardResult) AsSARIF(showDetails bool, logLevel zapcore.Level,
	writer io.Writer, checkDocs docs.Doc, minScore int) error {
	//nolint
	// https://docs.oasis-open.org/sarif/sarif/v2.1.0/cs01/sarif-v2.1.0-cs01.html.
	// We only support GitHub-supported properties:
	// see https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/sarif-support-for-code-scanning#supported-sarif-output-file-properties,
	// https://github.com/microsoft/sarif-tutorials.
	sarif := createSARIFHeader("https://github.com/ossf/scorecard",
		"supply-chain", "scorecard", r.Scorecard.Version, r.Scorecard.CommitSHA, r.Date)
	results := []result{}
	rules := []rule{}

	// nolint
	for i, check := range r.Checks {
		doc, e := checkDocs.GetCheck(check.Name)
		if e != nil {
			return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %v: %s", e, check.Name))
		}

		// Unclear what to use for PartialFingerprints.
		// GitHub only uses `primaryLocationLineHash`, which is not properly defined
		// and Appendix B of https://docs.oasis-open.org/sarif/sarif/v2.1.0/cs01/sarif-v2.1.0-cs01.html
		// warns about using line number for fingerprints:
		// "suppose the fingerprint were to include the line number where the result was located, and suppose
		// that after the baseline was constructed, a developer inserted additional lines of code above that
		// location. Then in the next run, the result would occur on a different line, the computed fingerprint
		// would change, and the result management system would erroneously report it as a new result."

		// Create locations and related locations.
		locs := detailsToLocations(check.Details2, showDetails, logLevel, minScore, check.Score)
		rlocs := detailsToRelatedLocations(check.Details2, showDetails, logLevel, minScore, check.Score)

		// Create check ID.
		//nolint:gosec
		m := md5.Sum([]byte(check.Name))
		checkID := hex.EncodeToString(m[:])

		// Create a header's rule.
		// TODO: verify `\n` is viewable in GitHub.
		rule := createSARIFRule(check.Name, checkID,
			doc.GetDocumentationURL(r.Scorecard.CommitSHA),
			doc.GetDescription(), doc.GetShort(),
			doc.GetTags())
		rules = append(rules, rule)

		// Add default location if no locations are present.
		// Note: GitHub needs at least one location to show the results.
		if len(locs) == 0 {
			locs = addDefaultLocation(locs)
		}

		// Create a result.
		r := createSARIFResult(i, checkID, check.Reason, minScore, check.Score, locs, rlocs)
		results = append(results, r)
	}

	// Set the results and rules to sarif.
	sarif.Runs[0].Tool.Driver.Rules = rules
	sarif.Runs[0].Results = results

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "   ")
	if err := encoder.Encode(sarif); err != nil {
		return sce.Create(sce.ErrScorecardInternal, err.Error())
	}

	return nil
}
