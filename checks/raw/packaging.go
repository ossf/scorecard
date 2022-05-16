// Copyright 2020 Security Scorecard Authors
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

package raw

import (
	"fmt"
	"path/filepath"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
)

// Packaging checks for packages.
func Packaging(c *checker.CheckRequest) (checker.PackagingData, error) {
	var data checker.PackagingData
	matchedFiles, err := c.RepoClient.ListFiles(fileparser.IsGithubWorkflowFileCb)
	if err != nil {
		return data, fmt.Errorf("%w", err)
	}
	if err != nil {
		return data, fmt.Errorf("RepoClient.ListFiles: %w", err)
	}

	for _, fp := range matchedFiles {
		fc, err := c.RepoClient.GetFileContent(fp)
		if err != nil {
			return data, fmt.Errorf("RepoClient.GetFileContent: %w", err)
		}

		workflow, errs := actionlint.Parse(fc)
		if len(errs) > 0 && workflow == nil {
			e := fileparser.FormatActionlintError(errs)
			return data, e
		}

		// Check if it's a packaging workflow.
		match, ok := isPackagingWorkflow(workflow, fp)
		// Always print debug messages.
		data.Packages = append(data.Packages,
			checker.Package{
				Msg: &match.Msg,
				File: &checker.File{
					Path:   fp,
					Type:   checker.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
			},
		)
		if !ok {
			continue
		}

		runs, err := c.RepoClient.ListSuccessfulWorkflowRuns(filepath.Base(fp))
		if err != nil {
			return data, fmt.Errorf("Client.Actions.ListWorkflowRunsByFileName: %w", err)
		}

		if len(runs) > 0 {
			// Create package.
			pkg := checker.Package{
				File: &checker.File{
					Path:   fp,
					Type:   checker.FileTypeSource,
					Offset: match.File.Offset,
				},
				Runs: []checker.Run{
					{
						URL: runs[0].URL,
					},
				},
			}
			// Create runs.
			for _, run := range runs {
				pkg.Runs = append(pkg.Runs,
					checker.Run{
						URL: run.URL,
					},
				)
			}
			data.Packages = append(data.Packages, pkg)

			return data, nil
		}

		data.Packages = append(data.Packages,
			checker.Package{
				// Debug message.
				Msg: stringPointer(fmt.Sprintf("GitHub publishing workflow not used in runs: %v", fp)),
				File: &checker.File{
					Path:   fp,
					Type:   checker.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
				// TODO: Job
			},
		)
	}

	// Return raw results.
	return data, nil
}

func stringPointer(s string) *string {
	return &s
}

// A packaging workflow.
func isPackagingWorkflow(workflow *actionlint.Workflow, fp string) (fileparser.JobMatchResult, bool) {
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

	return fileparser.RawAnyJobsMatch(workflow, jobMatchers, fp, "not a publishing workflow")
}
