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
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	"gopkg.in/yaml.v3"
)

var errNoScorecardYmlFile = errors.New("scorecard.yml doesn't exist or file doesn't contain exemptions")

// AnnotationReason is the reason behind an annotation.
type AnnotationReason string

const (
	// AnnotationTestData is to annotate when a check or probe is targeting a danger
	// in files or code snippets only used for test or example purposes.
	AnnotationTestData AnnotationReason = "test-data"
	// AnnotationRemediated is to annotate when a check or probe correctly identified a
	// danger and, even though the danger is necessary, a remediation was already applied.
	// E.g. a workflow is dangerous but only run under maintainers verification and approval,
	// or a binary is needed but it is signed or has provenance.
	AnnotationRemediated AnnotationReason = "remediated"
	// AnnotationNotApplicable is to annotate when a check or probe is not applicable for the case.
	// E.g. the dependencies should not be pinned because the project is a library.
	AnnotationNotApplicable AnnotationReason = "not-applicable"
	// AnnotationNotSupported is to annotate when the maintainer fulfills a check or probe in a way
	// that is not supported by Scorecard. E.g. Clang-Tidy is used as SAST tool but not identified
	// because its not supported.
	AnnotationNotSupported AnnotationReason = "not-supported"
	// AnnotationNotDetected is to annotate when the maintainer fulfills a check or probe in a way
	// that is supported by Scorecard but not identified. E.g. Dependabot is configured in the
	// repository settings and not in a file.
	AnnotationNotDetected AnnotationReason = "not-detected"
)

// AnnotationReasonDoc maps the reasons to their human-readable explanations.
var AnnotationReasonDoc = map[AnnotationReason]string{
	AnnotationTestData:      "The files or code snippets are only used for test or example purposes.",
	AnnotationRemediated:    "The dangerous files or code snippets are necessary but remediations were already applied.",
	AnnotationNotApplicable: "The check or probe is not applicable in this case.",
	AnnotationNotSupported:  "The check or probe is fulfilled but in a way that is not supported by Scorecard.",
	AnnotationNotDetected:   "The check or probe is fulfilled but in a way that is supported by Scorecard but it was not detected.",
}

// Annotation groups the annotation reason and, in the future, the related probe.
// If there is a probe, the reason applies to the probe.
// If there is not a probe, the reason applies to the check or group of checks in
// the exemption.
type Annotation struct {
	AnnotationReason AnnotationReason `yaml:"annotation"`
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

// parseScorecardYmlFile takes the scorecard.yml file content and returns a `MaintainersAnnotation`.
func parseScorecardYmlFile(ma *MaintainersAnnotation, content []byte) error {
	unmarshalErr := yaml.Unmarshal(content, ma)
	if unmarshalErr != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, unmarshalErr.Error())
	}

	return nil
}

// validateAndParseScorecardYmlFile takes the scorecard.yml files found, validates if its the correct
// configuration file and parse its configurations.
func validateAndParseScorecardYmlFile(path string, content []byte, args ...interface{}) (bool, error) {
	// Verify if received the necessary parameters to save the scorecard YML configurations
	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateAndParseScorecardYmlFile requires exactly one argument to save maintainers annotation data")
	}
	ma, ok := args[0].(*MaintainersAnnotation)
	if !ok {
		return false, fmt.Errorf(
			"validateAndParseScorecardYmlFile requires argument of type *MaintainersAnnotation")
	}

	// Exclude scorecard GitHub Action file
	if strings.Contains(path, ".github") {
		return true, nil
	}

	// Keep looking for files if content is empty
	if len(content) == 0 {
		return true, nil
	}

	// Found scorecard.yml file, parse file
	err := parseScorecardYmlFile(ma, content)
	if err != nil {
		return false, err
	}

	return false, nil
}

// readScorecardYmlFromRepo reads the scorecard.yml file from the repo and returns its configurations.
func readScorecardYmlFromRepo(repoClient clients.RepoClient) (MaintainersAnnotation, error) {
	ma := MaintainersAnnotation{}
	err := fileparser.OnMatchingFileContentDo(repoClient, fileparser.PathMatcher{
		Pattern:       "scorecard.yml",
		CaseSensitive: false,
	}, validateAndParseScorecardYmlFile, &ma)

	if err != nil {
		return ma, fmt.Errorf("%w", err)
	}

	if ma.Exemptions == nil {
		return ma, errNoScorecardYmlFile
	}

	return ma, nil
}

// GetMaintainersAnnotation reads the maintainers annotation from the repo, stored in scorecard.yml, and returns it.
func GetMaintainersAnnotation(repoClient clients.RepoClient) (MaintainersAnnotation, error) {
	// Read the scorecard.yml file and parse it into maintainers annotation
	ma, err := readScorecardYmlFromRepo(repoClient)

	// If scorecard.yml doesn't exist or file doesn't contain exemptions, ignore
	if err == errNoScorecardYmlFile {
		return MaintainersAnnotation{}, nil
	}
	// If an error happened while finding or parsing scorecard.yml file, then raise error
	if err != nil {
		return MaintainersAnnotation{}, fmt.Errorf("%w", err)
	}

	// Return maintainers annotation
	return ma, nil
}

// IsCheckExempted verifies if a given check in the results is exempted in maintainers annotation.
func (ma *MaintainersAnnotation) IsCheckExempted(check checker.CheckResult) (bool, []string) {
	// If check has a maximum score, then there it doesn't make sense anymore to reason the check
	// This may happen if the check score was once low but then the problem was fixed on Scorecard side
	// or on the maintainers side
	if check.Score == checker.MaxResultScore {
		return false, nil
	}

	for _, exemption := range ma.Exemptions {
		for _, checkName := range exemption.Checks {
			if strings.EqualFold(checkName, strings.ToLower(check.Name)) {
				return true, GetAnnotations(exemption.Annotations)
			}
		}
	}
	return false, nil
}

// GetAnnotations parses a group of annotations into annotation reasons.
func GetAnnotations(a []Annotation) []string {
	var annotationReasons []string
	for _, annotation := range a {
		annotationReasons = append(annotationReasons, AnnotationReasonDoc[annotation.AnnotationReason])
	}
	return annotationReasons
}
