// Copyright 2024 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package probes

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

type CheckName string

// Redefining check names here to avoid circular imports.
const (
	BinaryArtifacts      CheckName = "Binary-Artifacts"
	BranchProtection     CheckName = "Branch-Protection"
	CIIBestPractices     CheckName = "CII-Best-Practices"
	CITests              CheckName = "CI-Tests"
	CodeReview           CheckName = "Code-Review"
	Contributors         CheckName = "Contributors"
	DangerousWorkflow    CheckName = "Dangerous-Workflow"
	DependencyUpdateTool CheckName = "Dependency-Update-Tool"
	Fuzzing              CheckName = "Fuzzing"
	License              CheckName = "License"
	Maintained           CheckName = "Maintained"
	Packaging            CheckName = "Packaging"
	PinnedDependencies   CheckName = "Pinned-Dependencies"
	SAST                 CheckName = "SAST"
	SecurityPolicy       CheckName = "Security-Policy"
	SignedReleases       CheckName = "Signed-Releases"
	TokenPermissions     CheckName = "Token-Permissions"
	Vulnerabilities      CheckName = "Vulnerabilities"
	Webhooks             CheckName = "Webhooks"
)

type Probe struct {
	Name            string
	Implementation  ProbeImpl
	RequiredRawData []CheckName
}

type ProbeImpl func(*checker.RawResults) ([]finding.Finding, string, error)

// registered is the mapping of all registered probes.
var registered = map[string]Probe{}

func MustRegister(name string, impl ProbeImpl, requiredRawData []CheckName) {
	err := register(Probe{
		Name:            name,
		Implementation:  impl,
		RequiredRawData: requiredRawData,
	})
	if err != nil {
		panic(err)
	}
}

func register(p Probe) error {
	if p.Name == "" {
		return errors.CreateInternal(errors.ErrScorecardInternal, "name cannot be empty")
	}
	if p.Implementation == nil {
		return errors.CreateInternal(errors.ErrScorecardInternal, "implementation cannot be nil")
	}
	if len(p.RequiredRawData) == 0 {
		return errors.CreateInternal(errors.ErrScorecardInternal, "probes need some raw data")
	}
	registered[p.Name] = p
	return nil
}

func Get(name string) (Probe, error) {
	p, ok := registered[name]
	if !ok {
		return Probe{}, errors.CreateInternal(errors.ErrorUnsupportedCheck, "probe not found")
	}
	return p, nil
}
