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

package checknames

type CheckName = string

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
	SBOM                 CheckName = "SBOM"
	SecurityPolicy       CheckName = "Security-Policy"
	SignedReleases       CheckName = "Signed-Releases"
	TokenPermissions     CheckName = "Token-Permissions"
	Vulnerabilities      CheckName = "Vulnerabilities"
	Webhooks             CheckName = "Webhooks"
	MaintainerResponse   CheckName = "Maintainer-Response-BugSecurity"
)

var AllValidChecks []string = []string{
	BinaryArtifacts,
	BranchProtection,
	CIIBestPractices,
	CITests,
	CodeReview,
	Contributors,
	DangerousWorkflow,
	DependencyUpdateTool,
	Fuzzing,
	License,
	Maintained,
	Packaging,
	PinnedDependencies,
	SAST,
	SBOM,
	SecurityPolicy,
	SignedReleases,
	TokenPermissions,
	Vulnerabilities,
	Webhooks,
	MaintainerResponse,
}
