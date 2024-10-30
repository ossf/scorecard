// Copyright 2024 OpenSSF Scorecard Authors
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
package orgRequiresMFA

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const (
	Probe = "orgRequiresMFA"
)

func init() {
	// Register independently of any checks
	probes.MustRegisterIndependent(Probe, Run)
}

func Run(raw *checker.CheckRequest) (found []finding.Finding, probeName string, err error) {
	if raw == nil {
		err = fmt.Errorf("raw results is nil: %w", uerror.ErrNil)
		return found, Probe, err
	}

	mfaRequired, err := raw.RepoClient.GetMFARequired()
	if err != nil {
		err = fmt.Errorf("getting MFA required: %w", err)
		return found, Probe, err
	}

	var outcome finding.Outcome
	if mfaRequired {
		outcome = finding.OutcomeTrue
	} else {
		outcome = finding.OutcomeFalse
	}

	result, err := finding.NewWith(
		fs,
		Probe,
		"Collaborators require MFA",
		nil,
		outcome,
	)
	if err != nil {
		err = fmt.Errorf("creating finding: %w", err)
		return found, Probe, err
	}

	found = append(found, *result)

	return found, Probe, err
}
