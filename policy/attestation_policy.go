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
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/evaluation"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/remediation"
)

//nolint:govet
type AttestationPolicy struct {
	// PreventBinaryArtifacts : set to true to require that this project's SCM repo is
	// free of binary artifacts
	PreventBinaryArtifacts bool `yaml:"preventBinaryArtifacts"`

	// AllowedBinaryArtifacts : List of binary artifact paths to ignore
	// when checking for binary artifacts in a repo
	AllowedBinaryArtifacts []string `yaml:"allowedBinaryArtifacts"`

	// EnsureNoVulnerabilities : set to true to require that this project is free
	// of vulnerabilities, as discovered from the OSV service
	EnsureNoVulnerabilities bool `yaml:"ensureNoVulnerabilities"`

	// EnsurePinnedDependencies : set to true to require that this project pin dependencies
	// by hash/commit SHA
	EnsurePinnedDependencies bool `yaml:"ensurePinnedDependencies"`

	// EnsureCodeReviewed : set to true to require that the most recent commits in
	// this project have gone through a code review process
	EnsureCodeReviewed bool `yaml:"ensureCodeReviewed"`
}

type Dependency struct {
	Filepath    string `yaml:"filepath"`
	PackageName string `yaml:"packagename"`
	Version     string `yaml:"version"`
}

// Run attestation policy checks on raw data.
func RunChecksForPolicy(policy *AttestationPolicy, raw *checker.RawResults,
	dl checker.DetailLogger,
) (PolicyResult, error) {
	if policy.PreventBinaryArtifacts {
		checkResult, err := CheckPreventBinaryArtifacts(policy.AllowedBinaryArtifacts, raw, dl)

		if !checkResult || err != nil {
			return checkResult, err
		}
	}

	if policy.EnsureNoVulnerabilities {
		checkResult, err := CheckNoVulnerabilities(raw, dl)

		if !checkResult || err != nil {
			return checkResult, err
		}
	}

	if policy.EnsurePinnedDependencies {
		checkResult, err := CheckNoUnpinnedDependencies(raw, dl)

		if !checkResult || err != nil {
			return checkResult, err
		}
	}

	if policy.EnsureCodeReviewed {
		checkResult, err := CheckCodeReviewed(raw, dl)

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
			allowPath := allowedBinaryArtifacts[j]
			if allowPath != "" && strings.HasPrefix(artifactFile.Path, allowPath) {
				ignoreArtifact = true
				dl.Info(&checker.LogMessage{Text: fmt.Sprintf(
					"ignoring binary artifact at %s due to ignored path %s",
					artifactFile.Path,
					allowPath,
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

func CheckNoVulnerabilities(results *checker.RawResults, dl checker.DetailLogger) (PolicyResult, error) {
	nVulns := len(results.VulnerabilitiesResults.Vulnerabilities)

	if nVulns > 1 {
		dl.Info(&checker.LogMessage{Text: fmt.Sprintf("found %d vulnerabilities in package", nVulns)})
	}

	return nVulns > 0, nil
}

func CheckNoUnpinnedDependencies(results *checker.RawResults, dl checker.DetailLogger) (PolicyResult, error) {
	workflowPinning, pinningResults, err := evaluation.GetWorkflowPinningStatus(
		&results.PinningDependenciesResults,
		dl,
		remediation.RemediationMetadata{},
	)
	if err != nil {
		return Fail, fmt.Errorf("couldn't check workflow pinning status: %w", err)
	}

	if workflowPinning.ThirdParties == evaluation.NotPinned {
		dl.Info(&checker.LogMessage{Text: "third-party action workflow not pinned"})
		return Fail, nil
	}
	if workflowPinning.GitHubOwned == evaluation.NotPinned {
		dl.Info(&checker.LogMessage{Text: "github-owned action workflow not pinned"})
		return Fail, nil
	}

	for depType, pinningResult := range pinningResults {
		switch pinningResult {
		case evaluation.Pinned, evaluation.PinnedUndefined:
			dl.Debug(&checker.LogMessage{Text: fmt.Sprintf("%s dependencies pinned by hash", depType)})
		case evaluation.NotPinned:
			dl.Debug(&checker.LogMessage{Text: fmt.Sprintf("%s dependencies not pinned by hash", depType)})
			return Fail, nil
		}
	}

	return Pass, nil
}

func CheckCodeReviewed(results *checker.RawResults, dl checker.DetailLogger) (PolicyResult, error) {
	codeReviewResults := evaluation.CodeReview("", dl, &results.CodeReviewResults)

	return codeReviewResults.Score == 1, nil
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
