// Copyright 2023 OpenSSF Scorecard Authors
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

package probes

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	
	"github.com/ossf/scorecard/v4/probes/toolDependabotInstalled"
	"github.com/ossf/scorecard/v4/probes/toolPyUpInstalled"
	"github.com/ossf/scorecard/v4/probes/toolRenovateInstalled"
	"github.com/ossf/scorecard/v4/probes/toolSonarTypeLiftInstalled"
)

// ProbeImpl is the implementation of a probe.
type ProbeImpl func(*checker.RawResults) ([]finding.Finding, string, error)

var (
	// AllProbes represents all the probes.
	AllProbes []ProbeImpl
	// DependencyToolUpdates is all the probes for the
	// DpendencyUpdateTool check.
	DependencyToolUpdates = []ProbeImpl{
		toolRenovateInstalled.Run,
		toolDependabotInstalled.Run,
		toolPyUpInstalled.Run,
		toolSonarTypeLiftInstalled.Run,
	}
)

func concatMultipleSlices[T any](slices [][]T) []T {
	var totalLen int

	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, totalLen)

	var i int

	for _, s := range slices {
		i += copy(result[i:], s)
	}

	return result
}

func init() {
	AllProbes = concatMultipleSlices([][]ProbeImpl{
		DependencyToolUpdates,
	})
}
