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
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
)

type Probe struct {
	Name                      string
	Implementation            ProbeImpl
	IndependentImplementation IndependentProbeImpl
	RequiredRawData           []checknames.CheckName
}

type ProbeImpl func(*checker.RawResults) ([]finding.Finding, string, error)

type IndependentProbeImpl func(*checker.CheckRequest) ([]finding.Finding, string, error)

// registered is the mapping of all registered probes.
var registered = map[string]Probe{}

func MustRegister(name string, impl ProbeImpl, requiredRawData []checknames.CheckName) {
	err := register(Probe{
		Name:            name,
		Implementation:  impl,
		RequiredRawData: requiredRawData,
	})
	if err != nil {
		panic(err)
	}
}

func MustRegisterIndependent(name string, impl IndependentProbeImpl) {
	err := register(Probe{
		Name:                      name,
		IndependentImplementation: impl,
	})
	if err != nil {
		panic(err)
	}
}

func register(p Probe) error {
	if p.Name == "" {
		return errors.WithMessage(errors.ErrScorecardInternal, "name cannot be empty")
	}
	if p.Implementation == nil && p.IndependentImplementation == nil {
		return errors.WithMessage(errors.ErrScorecardInternal, "at least one implementation must be non-nil")
	}
	if p.Implementation != nil && len(p.RequiredRawData) == 0 {
		return errors.WithMessage(errors.ErrScorecardInternal, "non-independent probes need some raw data")
	}
	registered[p.Name] = p
	return nil
}

func Get(name string) (Probe, error) {
	p, ok := registered[name]
	if !ok {
		msg := fmt.Sprintf("probe %q not found", name)
		return Probe{}, errors.WithMessage(errors.ErrScorecardInternal, msg)
	}
	return p, nil
}
