// Copyright 2021 Security Scorecard Authors
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

package policy

import (
	"fmt"
	"os"

	"github.com/gobwas/glob"
	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

//nolint:govet
type AttestationPolicy struct {
	// PreventBinaryArtifacts : set to true to require that this project's SCM repo is
	// free of binary artifacts
	PreventBinaryArtifacts bool `yaml:"preventBinaryArtifacts"`

	// AllowedBinaryArtifacts : List of binary artifact paths to ignore
	// when checking for binary artifacts in a repo
	AllowedBinaryArtifacts []string `yaml:"allowedBinaryArtifacts"`
}

// Allows us to run fewer scorecard checks if some policy values
// are don't-cares
func (ap *AttestationPolicy) GetRequiredChecksForPolicy() map[string]bool {
	requiredChecks := make(map[string]bool)

	if ap.PreventBinaryArtifacts {
		requiredChecks["BinaryArtifacts"] = true
	}

	return requiredChecks
}

// Run attestation policy checks on raw data
func (ap *AttestationPolicy) EvaluateResults(raw *checker.RawResults) (PolicyResult, error) {
	dl := checker.NewLogger()
	if ap.PreventBinaryArtifacts {
		checkResult, err := CheckPreventBinaryArtifacts(ap.AllowedBinaryArtifacts, raw, dl)

		if !checkResult || err != nil {
			return checkResult, err
		}
	}

	return Pass, nil
}

type PolicyResult = bool

const (
	Pass PolicyResult = true
	Fail PolicyResult = false
)

func CheckPreventBinaryArtifacts(
	allowedBinaryArtifacts []string,
	results *checker.RawResults,
	dl checker.DetailLogger,
) (PolicyResult, error) {
	for i := range results.BinaryArtifactResults.Files {
		artifactFile := results.BinaryArtifactResults.Files[i]

		ignoreArtifact := false

		for j := range allowedBinaryArtifacts {
			// Treat user input as paths and try to match prefixes
			// This is a bit easier to use than forcing things to be file names
			allowGlob := allowedBinaryArtifacts[j]

			if g := glob.MustCompile(allowGlob); g.Match(artifactFile.Path) {
				ignoreArtifact = true
				dl.Info(&checker.LogMessage{Text: fmt.Sprintf(
					"ignoring binary artifact at %s due to ignored glob path %s",
					artifactFile.Path,
					g,
				)})
			}
		}

		if !ignoreArtifact {
			dl.Info(&checker.LogMessage{
				Path: artifactFile.Path, Type: checker.FileTypeBinary,
				Offset: artifactFile.Offset,
				Text:   "binary detected",
			})
			return Fail, nil
		}
	}

	return Pass, nil
}

// ParseFromFile takes a policy file and returns an AttestationPolicy.
func ParseAttestationPolicyFromFile(policyFile string) (*AttestationPolicy, error) {
	if policyFile != "" {
		data, err := os.ReadFile(policyFile)
		if err != nil {
			return nil, sce.WithMessage(sce.ErrScorecardInternal,
				fmt.Sprintf("os.ReadFile: %v", err))
		}

		sp, err := ParseAttestationPolicyFromYAML(data)
		if err != nil {
			return nil,
				sce.WithMessage(
					sce.ErrScorecardInternal,
					fmt.Sprintf("spol.ParseFromYAML: %v", err),
				)
		}

		return sp, nil
	}

	return nil, nil
}

// Parses a policy file and returns a AttestationPolicy.
func ParseAttestationPolicyFromYAML(b []byte) (*AttestationPolicy, error) {
	retPolicy := AttestationPolicy{}

	err := yaml.Unmarshal(b, &retPolicy)
	if err != nil {
		return &retPolicy, sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	return &retPolicy, nil
}
