package probes

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

type ProbeImpl func(*checker.RawResults) ([]finding.Finding, string, error)

// registered is the mapping of all registered probes.
var registered = map[string]ProbeImpl{}

// TODO: handle raw data and support types
func Register(name string, implementation ProbeImpl) error {
	if name == "" {
		//nolint:wrapcheck // fix this in config later
		return errors.CreateInternal(errors.ErrScorecardInternal, "name cannot be empty")
	}
	if implementation == nil {
		//nolint:wrapcheck // fix this in config later
		return errors.CreateInternal(errors.ErrScorecardInternal, "implementation cannot be nil")
	}
	registered[name] = implementation
	return nil
}
