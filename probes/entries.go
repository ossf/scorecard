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
	"github.com/ossf/scorecard/v4/probes/fuzzedWithClusterFuzzLite"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithGoNative"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithOSSFuzz"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithOneFuzz"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedHaskell"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedJavascript"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedTypescript"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsLinks"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsText"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsVulnerabilityDisclosure"
	"github.com/ossf/scorecard/v4/probes/securityPolicyPresent"
	"github.com/ossf/scorecard/v4/probes/toolDependabotInstalled"
	"github.com/ossf/scorecard/v4/probes/toolPyUpInstalled"
	"github.com/ossf/scorecard/v4/probes/toolRenovateInstalled"
	"github.com/ossf/scorecard/v4/probes/toolSonatypeLiftInstalled"
)

// ProbeImpl is the implementation of a probe.
type ProbeImpl func(*checker.RawResults) ([]finding.Finding, string, error)

var (
	// All represents all the probes.
	All []ProbeImpl
	// SecurityPolicy is all the probes for the
	// SecurityPolicy check.
	SecurityPolicy = []ProbeImpl{
		securityPolicyPresent.Run,
		securityPolicyContainsLinks.Run,
		securityPolicyContainsVulnerabilityDisclosure.Run,
		securityPolicyContainsText.Run,
	}
	// DependencyToolUpdates is all the probes for the
	// DpendencyUpdateTool check.
	DependencyToolUpdates = []ProbeImpl{
		toolRenovateInstalled.Run,
		toolDependabotInstalled.Run,
		toolPyUpInstalled.Run,
		toolSonatypeLiftInstalled.Run,
	}
	Fuzzing = []ProbeImpl{
		fuzzedWithOSSFuzz.Run,
		fuzzedWithOneFuzz.Run,
		fuzzedWithGoNative.Run,
		fuzzedWithClusterFuzzLite.Run,
		fuzzedWithPropertyBasedHaskell.Run,
		fuzzedWithPropertyBasedTypescript.Run,
		fuzzedWithPropertyBasedJavascript.Run,
	}
)

//nolint:gochecknoinits
func init() {
	All = concatMultipleProbes([][]ProbeImpl{
		DependencyToolUpdates,
		SecurityPolicy,
		Fuzzing,
	})
}

func concatMultipleProbes(slices [][]ProbeImpl) []ProbeImpl {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]ProbeImpl, 0, totalLen)
	for _, s := range slices {
		tmp = append(tmp, s...)
	}
	return tmp
}
