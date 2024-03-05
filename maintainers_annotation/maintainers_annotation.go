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

package maintainers_annotation

import (
	"errors"
	"fmt"
	"os"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	"gopkg.in/yaml.v3"
)

var errNoScorecardYmlFile = errors.New("scorecard.yml doesn't exist or file doesn't contain exemptions")

// Reason is the reason behind an annotation.
type Reason string

const (
	// TestData is to annotate when a check or probe is targeting a danger
	// in files or code snippets only used for test or example purposes.
	TestData Reason = "test-data"
	// Remediated is to annotate when a check or probe correctly identified a
	// danger and, even though the danger is necessary, a remediation was already applied.
	// E.g. a workflow is dangerous but only run under maintainers verification and approval,
	// or a binary is needed but it is signed or has provenance.
	Remediated Reason = "remediated"
	// NotApplicable is to annotate when a check or probe is not applicable for the case.
	// E.g. the dependencies should not be pinned because the project is a library.
	NotApplicable Reason = "not-applicable"
	// NotSupported is to annotate when the maintainer fulfills a check or probe in a way
	// that is not supported by Scorecard. E.g. Clang-Tidy is used as SAST tool but not identified
	// because its not supported.
	NotSupported Reason = "not-supported"
	// NotDetected is to annotate when the maintainer fulfills a check or probe in a way
	// that is supported by Scorecard but not identified. E.g. Dependabot is configured in the
	// repository settings and not in a file.
	NotDetected Reason = "not-detected"
)

// Annotation groups the annotation reason and, in the future, the related probe.
// If there is a probe, the reason applies to the probe.
// If there is not a probe, the reason applies to the check or group of checks in
// the exemption.
type Annotation struct {
	Reason Reason `yaml:"annotation"`
}

// Exemption defines a group of checks that are being annotated for various reasons.
type Exemption struct {
	Checks      []string     `yaml:"checks"`
	Annotations []Annotation `yaml:"annotations"`
}

// MaintainersAnnotation is a group of exemptions.
type MaintainersAnnotation struct {
	Exemptions []Exemption `yaml:"exemptions"`
}

// parseScorecardConfigFile takes the scorecard.yml file content and returns a `MaintainersAnnotation`.
func parseScorecardConfigFile(ma *MaintainersAnnotation, content []byte) error {
	unmarshalErr := yaml.Unmarshal(content, ma)
	if unmarshalErr != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, unmarshalErr.Error())
	}

	return nil
}

// readScorecardConfigFromRepo reads the scorecard.yml file from the repo and returns its configurations.
func readScorecardConfigFromRepo(repoClient clients.RepoClient) (MaintainersAnnotation, error) {
	ma := MaintainersAnnotation{}
	// Find scorecard.yml file in the repository's root
	content, err := repoClient.GetFileContent("scorecard.yml")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ma, errNoScorecardYmlFile
		}
		return ma, fmt.Errorf("fail to read scorecard configuration file: %w", err)
	}

	err = parseScorecardConfigFile(&ma, content)
	if err != nil {
		return ma, fmt.Errorf("fail to parse scorecard configuration file: %w", err)
	}

	if ma.Exemptions == nil {
		return ma, errNoScorecardYmlFile
	}

	return ma, nil
}

// GetMaintainersAnnotation reads the maintainers annotation from the repo, stored in scorecard.yml, and returns it.
func GetMaintainersAnnotation(repoClient clients.RepoClient) (MaintainersAnnotation, error) {
	// Read the scorecard.yml file and parse it into maintainers annotation
	ma, err := readScorecardConfigFromRepo(repoClient)

	// If scorecard.yml doesn't exist or file doesn't contain exemptions, ignore
	if err == errNoScorecardYmlFile {
		return MaintainersAnnotation{}, nil
	}
	// If an error happened while finding or parsing scorecard.yml file, then raise error
	if err != nil {
		return MaintainersAnnotation{}, fmt.Errorf("fail to get maintainers annotation: %w", err)
	}

	// Return maintainers annotation
	return ma, nil
}

// getAnnotations parses a group of annotations into annotation reasons.
func GetAnnotations(a []Annotation) []string {
	var reasons []string
	for _, annotation := range a {
		reasons = append(reasons, annotation.Reason.Doc())
	}
	return reasons
}

// Doc maps a reason to its human-readable explanation.
func (r *Reason) Doc() string {
	switch *r {
	case TestData:
		return "The files or code snippets are only used for test or example purposes."
	case Remediated:
		return "The dangerous files or code snippets are necessary but remediations were already applied."
	case NotApplicable:
		return "The check or probe is not applicable in this case."
	case NotSupported:
		return "The check or probe is fulfilled but in a way that is not supported by Scorecard."
	case NotDetected:
		return "The check or probe is fulfilled but in a way that is supported by Scorecard but it was not detected."
	default:
		return string(*r)
	}
}
