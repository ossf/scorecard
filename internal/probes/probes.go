package probes

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

type Probe struct {
	Name            string
	Implementation  ProbeImpl
	RequiredRawData []string
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
