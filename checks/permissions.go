// Copyright 2021 Security Scorecard Authors
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

package checks

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/evaluation"
	"github.com/ossf/scorecard/v4/checks/raw"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckTokenPermissions is the exported name for Token-Permissions check.
const CheckTokenPermissions = "Token-Permissions"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
		checker.CommitBased,
	}
	if err := registerCheck(CheckTokenPermissions, TokenPermissions, supportedRequestTypes); err != nil {
		// This should never happen.
		panic(err)
	}
}

// TokenPermissions will run the Token-Permissions check.
func TokenPermissions(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.TokenPermissions(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckTokenPermissions, e)
	}

	// Return raw results.
	if c.RawResults != nil {
		c.RawResults.TokenPermissionsResults = rawData
	}

	// Return the score evaluation.
	return evaluation.TokenPermissions(CheckTokenPermissions, c.Dlogger, &rawData)
}

// TODO: remove when migrated to raw results.
// Should be using the definition in raw/packaging.go.
func isPackagingWorkflow(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	jobMatchers := []fileparser.JobMatcher{
		{
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "actions/setup-node",
					With: map[string]string{"registry-url": "https://registry.npmjs.org"},
				},
				{
					Run: "npm.*publish",
				},
			},
			LogText: "candidate node publishing workflow using npm",
		},
		{
			// Java packages with maven.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "actions/setup-java",
				},
				{
					Run: "mvn.*deploy",
				},
			},
			LogText: "candidate java publishing workflow using maven",
		},
		{
			// Java packages with gradle.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "actions/setup-java",
				},
				{
					Run: "gradle.*publish",
				},
			},
			LogText: "candidate java publishing workflow using gradle",
		},
		{
			// Ruby packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: "gem.*push",
				},
			},
			LogText: "candidate ruby publishing workflow using gem",
		},
		{
			// NuGet packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: "nuget.*push",
				},
			},
			LogText: "candidate nuget publishing workflow",
		},
		{
			// Docker packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: "docker.*push",
				},
			},
			LogText: "candidate docker publishing workflow",
		},
		{
			// Docker packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "docker/build-push-action",
				},
			},
			LogText: "candidate docker publishing workflow",
		},
		{
			// Python packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "actions/setup-python",
				},
				{
					Uses: "pypa/gh-action-pypi-publish",
				},
			},
			LogText: "candidate python publishing workflow using pypi",
		},
		{
			// Python packages.
			// This is a custom Python packaging workflow based on semantic versioning.
			// TODO(#1642): accept custom workflows through a separate configuration.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "relekang/python-semantic-release",
				},
			},
			LogText: "candidate python publishing workflow using python-semantic-release",
		},
		{
			// Go packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "actions/setup-go",
				},
				{
					Uses: "goreleaser/goreleaser-action",
				},
			},
			LogText: "candidate golang publishing workflow",
		},
		{
			// Rust packages. https://doc.rust-lang.org/cargo/reference/publishing.html
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: "cargo.*publish",
				},
			},
			LogText: "candidate rust publishing workflow using cargo",
		},
	}

	return fileparser.AnyJobsMatch(workflow, jobMatchers, fp, dl, "not a publishing workflow")
}
