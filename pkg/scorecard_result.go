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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// ScorecardResult struct is returned on a successful Scorecard run.
type ScorecardResult struct {
	Repo     string
	Date     string
	Checks   []checker.CheckResult
	Metadata []string
}

//nolint
type jsonCheckResultV2 struct {
	Details []string
	Score   int
	Reason  string
	Name    string
}

type jsonScorecardResultV2 struct {
	Repo     string
	Date     string
	Checks   []jsonCheckResultV2
	Metadata []string
}

// AsJSON outputs the result in JSON format with a newline at the end.
// If called on []ScorecardResult will create NDJson formatted output.
// UPGRADEv2: will be removed.
func (r *ScorecardResult) AsJSON(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	if showDetails {
		if err := encoder.Encode(r); err != nil {
			//nolint:wrapcheck
			return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
		}
		return nil
	}
	out := ScorecardResult{
		Repo:     r.Repo,
		Date:     r.Date,
		Metadata: r.Metadata,
	}
	// UPGRADEv2: remove nolint after ugrade.
	//nolint
	for _, checkResult := range r.Checks {
		tmpResult := checker.CheckResult{
			Name:       checkResult.Name,
			Pass:       checkResult.Pass,
			Confidence: checkResult.Confidence,
		}
		out.Checks = append(out.Checks, tmpResult)
	}
	if err := encoder.Encode(out); err != nil {
		//nolint:wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}

// AsJSON2 exports results as JSON for new detail format.
func (r *ScorecardResult) AsJSON2(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)

	out := jsonScorecardResultV2{
		Repo:     r.Repo,
		Date:     r.Date,
		Metadata: r.Metadata,
	}

	//nolint
	for _, checkResult := range r.Checks {
		tmpResult := jsonCheckResultV2{
			Name:   checkResult.Name,
			Reason: checkResult.Reason,
			Score:  checkResult.Score,
		}
		if showDetails {
			for _, d := range checkResult.Details2 {
				tmpResult.Details = append(tmpResult.Details, d.Msg.Text)
			}
		}
		out.Checks = append(out.Checks, tmpResult)
	}
	if err := encoder.Encode(out); err != nil {
		//nolint:wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}

// AsCSV outputs ScorecardResult in CSV format.
func (r *ScorecardResult) AsCSV(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	w := csv.NewWriter(writer)
	record := []string{r.Repo}
	columns := []string{"Repository"}
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
		//nolint:wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("csv.Write: %v", err))
	}
	w.Flush()
	if err := w.Error(); err != nil {
		//nolint:wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("csv.Flush: %v", err))
	}
	return nil
}

// AsString returns ScorecardResult in string format.
func (r *ScorecardResult) AsString(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
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
			x[0] = fmt.Sprintf("%d", row.Score)
		}

		doc := fmt.Sprintf("github.com/ossf/scorecard/blob/main/docs/checks.md#%s", strings.ToLower(row.Name))
		x[1] = row.Reason
		x[2] = row.Name
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

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"Score", "Reason", "Name"}
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

func detailsToString(details []checker.CheckDetail, logLevel zapcore.Level) (string, bool) {
	// UPGRADEv2: change to make([]string, len(details))
	// followed by sa[i] = instead of append.
	var sa []string
	for _, v := range details {
		switch v.Msg.Version {
		//nolint
		case 3:
			if v.Type == checker.DetailDebug && logLevel != zapcore.DebugLevel {
				continue
			}
			if v.Msg.Path != "" {
				sa = append(sa, fmt.Sprintf("%s: %s: %s:%d", typeToString(v.Type), v.Msg.Text, v.Msg.Path, v.Msg.Offset))
			} else {
				sa = append(sa, fmt.Sprintf("%s: %s: %s", typeToString(v.Type), v.Msg.Text, v.Msg.Path))
			}
		default:
			if v.Type == checker.DetailDebug && logLevel != zapcore.DebugLevel {
				continue
			}
			sa = append(sa, fmt.Sprintf("%s: %s", typeToString(v.Type), v.Msg.Text))
		}
	}
	return strings.Join(sa, "\n"), len(sa) > 0
}

func typeToString(cd checker.DetailType) string {
	switch cd {
	default:
		panic("invalid detail")
	case checker.DetailInfo:
		return "Info"
	case checker.DetailWarn:
		return "Warn"
	case checker.DetailDebug:
		return "Debug"
	}
}
