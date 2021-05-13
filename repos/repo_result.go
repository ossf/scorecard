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

package repos

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ossf/scorecard/checker"
)

type RepoResult struct {
	Repo     string
	Date     string
	Checks   []checker.CheckResult
	Metadata []string
}

// AsJSON outputs the result in JSON format with a newline at the end.
// If called on []RepoResult will create NDJson formatted output.
func (r *RepoResult) AsJSON(showDetails bool, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	if showDetails {
		if err := encoder.Encode(r); err != nil {
			return fmt.Errorf("error encoding repo result as detailed JSON: %w", err)
		}
		return nil
	}
	out := RepoResult{
		Repo:     r.Repo,
		Date:     r.Date,
		Metadata: r.Metadata,
	}
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

func (r *RepoResult) AsCSV(showDetails bool, writer io.Writer) error {
	w := csv.NewWriter(writer)
	record := []string{r.Repo}
	columns := []string{"Repository"}
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

func (r *RepoResult) AsString(showDetails bool, writer io.Writer) error {
	fmt.Fprintf(writer, "Repo: %s\n", r.Repo)
	for _, checkResult := range r.Checks {
		fmt.Fprintf(writer, "%s: %s %d\n", checkResult.Name, displayResult(checkResult.Pass), checkResult.Confidence)
		if showDetails {
			for _, d := range checkResult.Details {
				fmt.Fprintf(writer, "%s\n", d)
			}
		}
	}
	return nil
}

func displayResult(result bool) string {
	if result {
		return "Pass"
	}
	return "Fail"
}
