// Copyright 2025 OpenSSF Scorecard Authors
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

package hasSelfHostedRunners

import (
	"embed"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

var errInvalidArg = errors.New("invalid arg")

const (
	Probe = "hasSelfHostedRunners"
	// selfHostedLabel is the label GitHub automatically applies to every self-hosted runner.
	// https://docs.github.com/en/actions/hosting-your-own-runners/using-self-hosted-runners-in-a-workflow
	selfHostedLabel = "self-hosted"
)

func init() {
	probes.MustRegisterIndependent(Probe, Run)
}

func Run(raw *checker.CheckRequest) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, Probe, fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	findings := []finding.Finding{}
	err := fileparser.OnMatchingFileContentDo(raw.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, checkSelfHostedRunners, &findings)
	if err != nil {
		return nil, Probe, fmt.Errorf("checking self-hosted runners: %w", err)
	}

	if len(findings) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"no self-hosted runners found in GitHub Actions workflows", nil, finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}

var checkSelfHostedRunners fileparser.DoWhileTrueOnFileContent = func(path string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}
	if len(args) != 1 {
		return false, fmt.Errorf("%w: expected 1 arg, got %d", errInvalidArg, len(args))
	}
	findings, ok := args[0].(*[]finding.Finding)
	if !ok {
		panic(fmt.Sprintf("expected *[]finding.Finding, got %v", reflect.TypeOf(args[0])))
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		f, err := finding.NewWith(fs, Probe, "malformed GitHub Actions workflow file",
			&finding.Location{Path: path}, finding.OutcomeError)
		if err != nil {
			return false, fmt.Errorf("create finding: %w", err)
		}
		*findings = append(*findings, *f)
		return true, nil
	}

	// Collect first so the findings for a single file are emitted in a stable order.
	// actionlint stores jobs in a map, whose iteration order is random.
	type flagged struct {
		id   string
		line uint
	}
	var jobs []flagged
	for _, job := range workflow.Jobs {
		pos := selfHostedRunnerPos(job)
		if pos == nil {
			continue
		}
		id := ""
		if job.ID != nil {
			id = job.ID.Value
		}
		jobs = append(jobs, flagged{id: id, line: uint(pos.Line)})
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].line < jobs[j].line
	})

	for i := range jobs {
		line := jobs[i].line
		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("job %q runs on a self-hosted runner", jobs[i].id),
			&finding.Location{Path: path, LineStart: &line}, finding.OutcomeTrue)
		if err != nil {
			return false, fmt.Errorf("create finding: %w", err)
		}
		*findings = append(*findings, *f)
	}

	return true, nil
}

// selfHostedRunnerPos reports the position of the runs-on element that marks the job
// as running on a self-hosted runner, or nil if the job uses a GitHub-hosted runner
// (or can't be resolved statically).
func selfHostedRunnerPos(job *actionlint.Job) *actionlint.Pos {
	if job == nil || job.RunsOn == nil {
		return nil
	}
	// Runner groups only exist for self-hosted runners.
	if job.RunsOn.Group != nil && job.RunsOn.Group.Value != "" {
		return job.RunsOn.Group.Pos
	}
	for _, label := range job.RunsOn.Labels {
		if label != nil && strings.EqualFold(label.Value, selfHostedLabel) {
			return label.Pos
		}
	}
	return nil
}
