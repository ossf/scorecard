// Copyright 2020 Security Scorecard Authors
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
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v2/checker"
	docs "github.com/ossf/scorecard/v2/docs/checks"
	sce "github.com/ossf/scorecard/v2/errors"
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

// ScorecardResult struct is returned on a successful Scorecard run.
type ScorecardResult struct {
	Repo      RepoInfo
	Date      time.Time
	Scorecard ScorecardInfo
	Checks    []checker.CheckResult
	Metadata  []string
}

// AsCSV outputs ScorecardResult in CSV format.
func (r *ScorecardResult) AsCSV(showDetails bool, logLevel zapcore.Level,
	checkDocs docs.Doc, writer io.Writer) error {
	score, err := r.aggregateScore(checkDocs)
	if err != nil {
		return err
	}
	w := csv.NewWriter(writer)
	record := []string{r.Repo.Name, fmt.Sprintf("%.1f", score)}
	columns := []string{"Repository", "AggScore"}
	// UPGRADEv2: remove nolint after ugrade.
	//nolint
	for _, checkResult := range r.Checks {
		columns = append(columns, checkResult.Name+"_Pass", checkResult.Name+"_Confidence")
		record = append(record, strconv.FormatBool(checkResult.Pass),
			strconv.Itoa(checkResult.Confidence))
		if showDetails {
			columns = append(columns, checkResult.Name+"_Details")
			record = append(record, checkResult.Details...)
		}
	}
	fmt.Fprintf(writer, "%s\n", strings.Join(columns, ","))
	if err := w.Write(record); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("csv.Write: %v", err))
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("csv.Flush: %v", err))
	}
	return nil
}

func (r *ScorecardResult) aggregateScore(checkDocs docs.Doc) (float64, error) {
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
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Invalid risk for %s: %s", check.Name, risk))
		}

		total += rs
		score += rs * float64(check.Score)
	}

	return score / total, nil
}

// AsString returns ScorecardResult in string format.
func (r *ScorecardResult) AsString(showDetails bool, logLevel zapcore.Level,
	checkDocs docs.Doc, writer io.Writer) error {
	data := make([][]string, len(r.Checks))
	//nolint
	for i, row := range r.Checks {
		const withdetails = 5
		const withoutdetails = 4
		var x []string

		if showDetails {
			x = make([]string, withdetails)
		} else {
			x = make([]string, withoutdetails)
		}

		// UPGRADEv2: rename variable.
		if row.Score == checker.InconclusiveResultScore {
			x[0] = "?"
		} else {
			x[0] = fmt.Sprintf("%d / %d", row.Score, checker.MaxResultScore)
		}

		cdoc, e := checkDocs.GetCheck(row.Name)
		if e != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %s: %v", row.Name, e))
		}

		doc := cdoc.GetDocumentationURL(r.Scorecard.CommitSHA)
		x[1] = row.Name
		x[2] = row.Reason
		if showDetails {
			details, show := detailsToString(row.Details2, logLevel)
			if show {
				x[3] = details
			}
			x[4] = doc
		} else {
			x[3] = doc
		}

		data[i] = x
	}

	score, err := r.aggregateScore(checkDocs)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Aggregate score: %.1f / %d\n\n", score, checker.MaxResultScore)
	fmt.Fprintln(os.Stdout, "Check scores:")

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"Score", "Name", "Reason"}
	if showDetails {
		header = append(header, "Details")
	}
	header = append(header, "Documentation/Remediation")
	table.SetHeader(header)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetRowSeparator("-")
	table.SetRowLine(true)
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(true)
	table.Render()

	return nil
}
