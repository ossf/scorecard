package probes

import (
	"github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/probes"
)

// registered is the mapping of all registered probes.
var registered = map[string]probes.ProbeImpl{}

// TODO: handle raw data and support types
func Register(name string, implementation probes.ProbeImpl) error {
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
