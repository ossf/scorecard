// Copyright 2026 OpenSSF Scorecard Authors
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
	"slices"
	"testing"

	"github.com/ossf/scorecard/v5/internal/checknames"
	internalprobes "github.com/ossf/scorecard/v5/internal/probes"
)

func TestTokenPermissionProbesRegistered(t *testing.T) {
	t.Parallel()

	probeNames := []string{
		"jobLevelPermissions",
		"topLevelPermissions",
		"hasNoGitHubWorkflowPermissionUnknown",
	}

	for _, name := range probeNames {
		p, err := internalprobes.Get(name)
		if err != nil {
			t.Errorf("probe %q is not registered: %v", name, err)
			continue
		}
		if p.Implementation == nil {
			t.Errorf("probe %q has nil Implementation", name)
		}
		if !slices.Contains(p.RequiredRawData, checknames.TokenPermissions) {
			t.Errorf("probe %q does not declare %s as required raw data", name, checknames.TokenPermissions)
		}
	}
}
