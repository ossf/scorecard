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
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/checker"
)

type ScorecardResult struct {
	Repo     string
	Date     string
	Checks   []checker.CheckResult
	Metadata []string
}

// AsJSON outputs the result in JSON format with a newline at the end.
// If called on []ScorecardResult will create NDJson formatted output.
func (r *ScorecardResult) AsJSON(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	if showDetails {
		if err := encoder.Encode(r); err != nil {
			return fmt.Errorf("error encoding repo result as detailed JSON: %w", err)
		}
		return nil
	}
	out := ScorecardResult{
		Repo:     r.Repo,
		Date:     r.Date,
		Metadata: r.Metadata,
	}
	// UPGRADEv2: remove nolint after uggrade.
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
		return fmt.Errorf("error encoding repo result as JSON: %w", err)
	}
	return nil
}

func (r *ScorecardResult) AsCSV(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	w := csv.NewWriter(writer)
	record := []string{r.Repo}
	columns := []string{"Repository"}
	// UPGRADEv2: remove nolint after uggrade.
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
		return fmt.Errorf("error writing repo result as CSV: %w", err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("error flushing repo result as CSV: %w", err)
	}
	return nil
}

// UPGRADEv2: will be removed.
func (r *ScorecardResult) AsString(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	sortedChecks := make([]checker.CheckResult, len(r.Checks))
	//nolint
	for i, checkResult := range r.Checks {
		sortedChecks[i] = checkResult
	}
	sort.Slice(sortedChecks, func(i, j int) bool {
		if sortedChecks[i].Pass == sortedChecks[j].Pass {
			return sortedChecks[i].Name < sortedChecks[j].Name
		}
		return sortedChecks[i].Pass
	})

	data := make([][]string, len(sortedChecks))
	//nolint
	for i, row := range sortedChecks {
		const withdetails = 4
		const withoutdetails = 3
		var x []string

		if showDetails {
			x = make([]string, withdetails)
		} else {
			x = make([]string, withoutdetails)
		}

		x[0] = displayResult(row.Pass)
		x[1] = strconv.Itoa(row.Confidence)
		x[2] = row.Name
		if showDetails {
			//nolint
			if row.Version == 2 {
				details, show := detailsToString(row.Details2, logLevel)
				if !show {
					continue
				}
				x[3] = details
			} else {
				x[3] = strings.Join(row.Details, "\n")
			}
		}
		data[i] = x
	}

	fmt.Fprintf(writer, "Repo: %s\n", r.Repo)
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"Status", "Confidence", "Name"}
	if showDetails {
		header = append(header, "Details")
	}
	table.SetHeader(header)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetRowSeparator("-")
	table.SetRowLine(true)
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()

	return nil
}

// UPGRADEv2: new code.
func (r *ScorecardResult) AsString2(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	sortedChecks := make([]checker.CheckResult, len(r.Checks))
	//nolint
	// UPGRADEv2: not needed after upgrade.
	for i, checkResult := range r.Checks {
		sortedChecks[i] = checkResult
	}
	sort.Slice(sortedChecks, func(i, j int) bool {
		if sortedChecks[i].Pass == sortedChecks[j].Pass {
			return sortedChecks[i].Name < sortedChecks[j].Name
		}
		return sortedChecks[i].Pass
	})

	data := make([][]string, len(sortedChecks))
	//nolint
	// UPGRADEv2: not needed after upgrade.
	for i, row := range sortedChecks {
		//nolint
		if row.Version != 2 {
			continue
		}
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

		doc := fmt.Sprintf("https://github.com/ossf/scorecard/blob/main/checks/checks.md#%s", strings.ToLower(row.Name))
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
	//nolint
	var sa []string
	for _, v := range details {
		if v.Type == checker.DetailDebug && logLevel != zapcore.DebugLevel {
			continue
		}
		sa = append(sa, fmt.Sprintf("%s: %s", typeToString(v.Type), v.Msg))
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

// UPGRADEv2: not needed after upgrade.
func displayResult(result bool) string {
	if result {
		return "Pass"
	}
	return "Fail"
}
