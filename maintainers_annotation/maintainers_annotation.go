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

	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	"gopkg.in/yaml.v3"
)

var errNoScorecardYmlFile = errors.New("scorecard.yml doesn't exist or file doesn't contain exemptions")

// AnnotationReason is the reason behind an annotation.
type AnnotationReason string

const (
	// AnnotationTestData is to annotate when the target is used only for test purposes.
	AnnotationTestData AnnotationReason = "test-data"
	// AnnotationSignedBinaries is to annotate when the targeted binary is signed.
	AnnotationSignedBinaries AnnotationReason = "signed-binaries"

	// TODO: Complete annotation reasons
)

// Annotation groups the annotation reason and, in the future, the related probe.
// If there is a probe, the reason applies to the probe.
// If there is not a probe, the reason applies to the check or group of checks in
// the exemption.
type Annotation struct {
	Annotation AnnotationReason `yaml:"annotation"`
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

// TODO:
// Identify which checks are annotated.
// Do not upload these checks to GitHub security alerts.
