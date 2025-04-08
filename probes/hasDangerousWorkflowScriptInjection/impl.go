// Copyright 2023 OpenSSF Scorecard Authors
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

//nolint:stylecheck
package hasDangerousWorkflowScriptInjection

import (
	"embed"
	"fmt"
	"os"
	"path"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/hasDangerousWorkflowScriptInjection/internal/patch"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.DangerousWorkflow})
}

//go:embed *.yml
var fs embed.FS

const Probe = "hasDangerousWorkflowScriptInjection"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.DangerousWorkflowResults

	if r.NumWorkflows == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Project does not have any workflows.", nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}

	var findings []finding.Finding
	var currWorkflow string
	var workflow *actionlint.Workflow
	var content []byte
	var errs []*actionlint.Error
	localPath := raw.Metadata.Metadata["localPath"]
	for _, w := range r.Workflows {
		if w.Type != checker.DangerousWorkflowScriptInjection {
			continue
		}

		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("script injection with untrusted input '%v'", w.File.Snippet),
			nil, finding.OutcomeTrue)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		f = f.WithLocation(&finding.Location{
			Path:      w.File.Path,
			Type:      w.File.Type,
			LineStart: &w.File.Offset,
			Snippet:   &w.File.Snippet,
		})

		err = parseWorkflow(localPath, &w, &currWorkflow, &content, &workflow, &errs)
		if err == nil {
			generatePatch(&w, content, workflow, errs, f)
		}

		findings = append(findings, *f)
	}

	if len(findings) == 0 {
		return falseOutcome()
	}

	return findings, Probe, nil
}

func parseWorkflow(
	localPath string,
	e *checker.DangerousWorkflow,
	currWorkflow *string,
	content *[]byte,
	workflow **actionlint.Workflow,
	errs *[]*actionlint.Error,
) error {
	var err error
	wp := path.Join(localPath, e.File.Path)
	if *currWorkflow != wp {
		// update current open file if injection in different file
		*currWorkflow = wp
		*content, err = os.ReadFile(wp)
		if err != nil {
			return err //nolint:wrapcheck // we only care about the error's existence
		}

		*workflow, *errs = actionlint.Parse(*content)
		if len(*errs) > 0 && *workflow == nil {
			// the workflow contains unrecoverable parsing errors, skip.
			return err //nolint:wrapcheck // we only care about the error's existence
		}
	}
	return nil
}

func generatePatch(
	e *checker.DangerousWorkflow,
	content []byte,
	workflow *actionlint.Workflow,
	errs []*actionlint.Error,
	f *finding.Finding,
) {
	findingPatch, err := patch.GeneratePatch(e.File, content, workflow, errs)
	if err != nil {
		return
	}
	f.WithPatch(&findingPatch)
	f.WithRemediationMetadata(map[string]string{
		"patch": findingPatch,
	})
}

func falseOutcome() ([]finding.Finding, string, error) {
	f, err := finding.NewWith(fs, Probe,
		"Project does not have dangerous workflow(s) with possibility of script injection.", nil,
		finding.OutcomeFalse)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	return []finding.Finding{*f}, Probe, nil
}
