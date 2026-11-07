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

package checkdef

import (
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/evaluation"
	"github.com/ossf/scorecard/v5/checks/raw"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/probes"
)

// BinaryArtifacts is the Binary-Artifacts check.
var BinaryArtifacts = define(
	checknames.BinaryArtifacts,
	raw.BinaryArtifacts,
	func(r *checker.RawResults, data checker.BinaryArtifactData) { r.BinaryArtifactResults = data },
	probes.BinaryArtifacts,
	evaluation.BinaryArtifacts,
)

// BranchProtection is the Branch-Protection check.
var BranchProtection = define(
	checknames.BranchProtection,
	raw.BranchProtection,
	func(r *checker.RawResults, data checker.BranchProtectionsData) { r.BranchProtectionResults = data },
	probes.BranchProtection,
	evaluation.BranchProtection,
)

// CIIBestPractices is the CII-Best-Practices check.
var CIIBestPractices = define(
	checknames.CIIBestPractices,
	raw.CIIBestPractices,
	func(r *checker.RawResults, data checker.CIIBestPracticesData) { r.CIIBestPracticesResults = data },
	probes.CIIBestPractices,
	evaluation.CIIBestPractices,
)

// CITests is the CI-Tests check.
var CITests = define(
	checknames.CITests,
	func(c *checker.CheckRequest) (checker.CITestData, error) { return raw.CITests(c.RepoClient) },
	func(r *checker.RawResults, data checker.CITestData) { r.CITestResults = data },
	probes.CITests,
	evaluation.CITests,
)

// Contributors is the Contributors check.
var Contributors = define(
	checknames.Contributors,
	raw.Contributors,
	func(r *checker.RawResults, data checker.ContributorsData) { r.ContributorsResults = data },
	probes.Contributors,
	evaluation.Contributors,
)

// DangerousWorkflow is the Dangerous-Workflow check.
var DangerousWorkflow = define(
	checknames.DangerousWorkflow,
	raw.DangerousWorkflow,
	func(r *checker.RawResults, data checker.DangerousWorkflowData) { r.DangerousWorkflowResults = data },
	probes.DangerousWorkflows,
	evaluation.DangerousWorkflow,
)

// Maintained is the Maintained check.
var Maintained = define(
	checknames.Maintained,
	raw.Maintained,
	func(r *checker.RawResults, data checker.MaintainedData) { r.MaintainedResults = data },
	probes.Maintained,
	evaluation.Maintained,
)
