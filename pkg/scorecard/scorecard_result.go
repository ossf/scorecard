// Copyright 2020 OpenSSF Scorecard Authors
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

package scorecard

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks"
	"github.com/ossf/scorecard/v5/checks/raw"
	"github.com/ossf/scorecard/v5/checks/raw/github"
	"github.com/ossf/scorecard/v5/checks/raw/gitlab"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/config"
	docChecks "github.com/ossf/scorecard/v5/docs/checks"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	proberegistration "github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
	spol "github.com/ossf/scorecard/v5/policy"
)

// ScorecardInfo contains information about the scorecard code that was run.
type ScorecardInfo struct {
	Version   string
	CommitSHA string
}

// RepoInfo contains information about the repo that was analyzed.
type RepoInfo struct {
	Name      string
	CommitSHA string
}

// Result struct is returned on a successful Scorecard run.
type Result struct {
	Repo       RepoInfo
	Date       time.Time
	Scorecard  ScorecardInfo
	Checks     []checker.CheckResult
	RawResults checker.RawResults
	Findings   []finding.Finding
	Metadata   []string
	Config     config.Config
}

// AsStringResultOption provides configuration options for string Scorecard results.
type AsStringResultOption struct {
	LogLevel    log.Level
	Details     bool
	Annotations bool
}

func scoreToString(s float64) string {
	if s == checker.InconclusiveResultScore {
		return "?"
	}
	return fmt.Sprintf("%.1f", s)
}

// GetAggregateScore returns the aggregate score.
func (r *Result) GetAggregateScore(checkDocs docChecks.Doc) (float64, error) {
	// TODO: calculate the score and make it a field
	// of ScorecardResult
	weights := map[string]float64{"Critical": 10, "High": 7.5, "Medium": 5, "Low": 2.5}
	// Note: aggregate score changes depending on which checks are run.
	total := float64(0)
	score := float64(0)
	for i := range r.Checks {
		check := r.Checks[i]
		doc, e := checkDocs.GetCheck(check.Name)
		if e != nil {
			return checker.InconclusiveResultScore,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %s: %v", check.Name, e))
		}

		risk := doc.GetRisk()
		rs, exists := weights[risk]
		if !exists {
			return checker.InconclusiveResultScore,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Invalid risk for %s: '%s'", check.Name, risk))
		}

		// This indicates an inconclusive score.
		if check.Score < checker.MinResultScore {
			continue
		}

		total += rs
		score += rs * float64(check.Score)
	}

	// Inconclusive result.
	if total == 0 {
		return checker.InconclusiveResultScore, nil
	}

	return score / total, nil
}

// FormatResults formats scorecard results.
func FormatResults(
	opts *options.Options,
	results *Result,
	doc docChecks.Doc,
	policy *spol.ScorecardPolicy,
) error {
	var err error

	// Define output to console or file
	output := os.Stdout
	if opts.ResultsFile != "" {
		output, err = os.Create(opts.ResultsFile)
		if err != nil {
			return fmt.Errorf("unable to create output file: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Writing results to", opts.ResultsFile)
		defer output.Close()
	}

	switch opts.Format {
	case options.FormatDefault:
		o := &AsStringResultOption{
			Details:     opts.ShowDetails,
			Annotations: opts.ShowAnnotations,
			LogLevel:    log.ParseLevel(opts.LogLevel),
		}
		err = results.AsString(output, doc, o)
	case options.FormatSarif:
		// TODO: support config files and update checker.MaxResultScore.
		err = results.AsSARIF(opts.ShowDetails, log.ParseLevel(opts.LogLevel), output, doc, policy, opts)
	case options.FormatJSON:
		o := &AsJSON2ResultOption{
			Details:     opts.ShowDetails,
			Annotations: opts.ShowAnnotations,
			LogLevel:    log.ParseLevel(opts.LogLevel),
		}
		err = results.AsJSON2(output, doc, o)
	case options.FormatInToto:
		o := &AsInTotoResultOption{
			AsJSON2ResultOption: AsJSON2ResultOption{
				Details:     opts.ShowDetails,
				Annotations: opts.ShowAnnotations,
				LogLevel:    log.ParseLevel(opts.LogLevel),
			},
		}
		err = results.AsInToto(output, doc, o)
	case options.FormatProbe:
		var opts *ProbeResultOption
		err = results.AsProbe(output, opts)
	case options.FormatRaw:
		err = results.AsRawJSON(output)
	default:
		err = sce.WithMessage(
			sce.ErrScorecardInternal,
			fmt.Sprintf(
				"invalid format flag: %v. Expected [default, json]",
				opts.Format,
			),
		)
	}

	if err != nil {
		return fmt.Errorf("failed to output results: %w", err)
	}

	return nil
}

// AsString returns ScorecardResult in string format.
func (r *Result) AsString(writer io.Writer, checkDocs docChecks.Doc, opt *AsStringResultOption) error {
	if opt == nil {
		opt = &AsStringResultOption{
			LogLevel:    log.DefaultLevel,
			Details:     false,
			Annotations: false,
		}
	}

	data := make([][]string, len(r.Checks))

	for i, row := range r.Checks {
		var x []string

		// UPGRADEv2: rename variable.
		if row.Score == checker.InconclusiveResultScore {
			x = append(x, "?")
		} else {
			x = append(x, fmt.Sprintf("%d / %d", row.Score, checker.MaxResultScore))
		}

		cdoc, e := checkDocs.GetCheck(row.Name)
		if e != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %s: %v", row.Name, e))
		}

		doc := cdoc.GetDocumentationURL(r.Scorecard.CommitSHA)
		x = append(x, row.Name, row.Reason)
		if opt.Details {
			details, show := detailsToString(row.Details, opt.LogLevel)
			if show {
				x = append(x, details)
			}
		}
		x = append(x, doc)
		if opt.Annotations {
			reasons := row.Annotations(r.Config)
			x = append(x, strings.Join(reasons, "\n"))
		}
		data[i] = x
	}

	score, err := r.GetAggregateScore(checkDocs)
	if err != nil {
		return err
	}
	s := fmt.Sprintf("Aggregate score: %s / %d\n\n", scoreToString(score), checker.MaxResultScore)
	if score == checker.InconclusiveResultScore {
		s = "Aggregate score: ?\n\n"
	}
	fmt.Fprint(writer, s)
	fmt.Fprintln(writer, "Check scores:")

	cfg := tablewriter.Config{
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
		},
	}
	rendition := tw.Rendition{
		Symbols: tw.NewSymbolCustom("scorecard table legacy").WithCenter("|").
			WithBottomLeft("|").WithBottomRight("|").
			WithMidLeft("|").WithMidRight("|").
			WithTopLeft("|").WithTopRight("|"),
	}
	table := tablewriter.NewTable(writer, tablewriter.WithConfig(cfg), tablewriter.WithRendition(rendition))
	header := []string{"Score", "Name", "Reason"}
	if opt.Details {
		header = append(header, "Details")
	}
	header = append(header, "Documentation/Remediation")
	if opt.Annotations {
		header = append(header, "Annotation")
	}
	table.Header(header)
	if err := table.Bulk(data); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("tablewriter Bulk: %v", err))
	}
	if err := table.Render(); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("tablewriter Render: %v", err))
	}

	return nil
}

//nolint:gocognit,gocyclo // nothing better to do right now
func assignRawData(probeCheckName string, request *checker.CheckRequest, ret *Result) error {
	switch probeCheckName {
	case checks.CheckBinaryArtifacts:
		rawData, err := raw.BinaryArtifacts(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.BinaryArtifactResults = rawData
	case checks.CheckBranchProtection:
		rawData, err := raw.BranchProtection(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.BranchProtectionResults = rawData
	case checks.CheckCIIBestPractices:
		rawData, err := raw.CIIBestPractices(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.CIIBestPracticesResults = rawData
	case checks.CheckCITests:
		rawData, err := raw.CITests(request.RepoClient)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.CITestResults = rawData
	case checks.CheckCodeReview:
		rawData, err := raw.CodeReview(request.RepoClient)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.CodeReviewResults = rawData
	case checks.CheckContributors:
		rawData, err := raw.Contributors(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.ContributorsResults = rawData
	case checks.CheckDangerousWorkflow:
		rawData, err := raw.DangerousWorkflow(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.DangerousWorkflowResults = rawData
	case checks.CheckDependencyUpdateTool:
		rawData, err := raw.DependencyUpdateTool(request.RepoClient)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.DependencyUpdateToolResults = rawData
	case checks.CheckFuzzing:
		rawData, err := raw.Fuzzing(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.FuzzingResults = rawData
	case checks.CheckLicense:
		rawData, err := raw.License(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.LicenseResults = rawData
	case checks.CheckMaintained:
		rawData, err := raw.Maintained(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.MaintainedResults = rawData
	case checks.CheckPackaging:
		switch request.RepoClient.(type) {
		case *githubrepo.Client:
			rawData, err := github.Packaging(request)
			if err != nil {
				return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			}
			ret.RawResults.PackagingResults = rawData
		case *gitlabrepo.Client:
			rawData, err := gitlab.Packaging(request)
			if err != nil {
				return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			}
			ret.RawResults.PackagingResults = rawData
		default:
			return sce.WithMessage(sce.ErrScorecardInternal, "Only github and gitlab are supported")
		}
	case checks.CheckPinnedDependencies:
		rawData, err := raw.PinningDependencies(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.PinningDependenciesResults = rawData
	case checks.CheckSAST:
		rawData, err := raw.SAST(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.SASTResults = rawData
	case checks.CheckSBOM:
		rawData, err := raw.SBOM(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.SBOMResults = rawData
	case checks.CheckSecurityPolicy:
		rawData, err := raw.SecurityPolicy(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.SecurityPolicyResults = rawData
	case checks.CheckSignedReleases:
		rawData, err := raw.SignedReleases(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.SignedReleasesResults = rawData
	case checks.CheckTokenPermissions:
		rawData, err := raw.TokenPermissions(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.TokenPermissionsResults = rawData
	case checks.CheckVulnerabilities:
		rawData, err := raw.Vulnerabilities(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.VulnerabilitiesResults = rawData
	case checks.CheckWebHooks:
		rawData, err := raw.WebHook(request)
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		ret.RawResults.WebhookResults = rawData
	default:
		return sce.WithMessage(sce.ErrScorecardInternal, "unknown check")
	}
	return nil
}

func populateRawResults(request *checker.CheckRequest, probesToRun []string, ret *Result) error {
	seen := map[string]bool{}
	for _, probeName := range probesToRun {
		p, err := proberegistration.Get(probeName)
		if err != nil {
			return fmt.Errorf("getting probe %q: %w", probeName, err)
		}
		for _, checkName := range p.RequiredRawData {
			if !seen[checkName] {
				err := assignRawData(checkName, request, ret)
				if err != nil {
					return err
				}
				seen[checkName] = true
			}
		}
	}
	return nil
}
