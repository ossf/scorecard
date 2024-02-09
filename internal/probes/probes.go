package probes

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

type CheckName string

// Redefining check names here to avoid circular imports
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
)

type Probe struct {
	Name            string
	Implementation  ProbeImpl
	RequiredRawData []CheckName
}

type ProbeImpl func(*checker.RawResults) ([]finding.Finding, string, error)

// registered is the mapping of all registered probes.
var registered = map[string]Probe{}

// TODO: handle raw data and support types
func Register(probe Probe) error {
	if probe.Name == "" {
		//nolint:wrapcheck // fix this in config later
		return errors.CreateInternal(errors.ErrScorecardInternal, "name cannot be empty")
	}
	if probe.Implementation == nil {
		//nolint:wrapcheck // fix this in config later
		return errors.CreateInternal(errors.ErrScorecardInternal, "implementation cannot be nil")
	}
	registered[probe.Name] = probe
	return nil
}

func Get(name string) (Probe, error) {
	p, ok := registered[name]
	if !ok {
		return Probe{}, errors.CreateInternal(errors.ErrorUnsupportedCheck, "probe not found")
	}
	return p, nil
}
