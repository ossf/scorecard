// Copyright 2020 OpenSSF Scorecard Authors
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
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
)

const (
	dependabotID = 49699333
)

// DependencyUpdateTool is the exported name for Depdendency-Update-Tool.
func DependencyUpdateTool(c clients.RepoClient) (checker.DependencyUpdateToolData, error) {
	var tools []checker.Tool
	err := fileparser.OnAllFilesDo(c, checkDependencyFileExists, &tools)
	if err != nil {
		return checker.DependencyUpdateToolData{}, fmt.Errorf("%w", err)
	}

	if len(tools) != 0 {
		return checker.DependencyUpdateToolData{Tools: tools}, nil
	}

	commits, err := c.SearchCommits(clients.SearchCommitsOptions{Author: "dependabot[bot]"})
	if err != nil {
		return checker.DependencyUpdateToolData{}, fmt.Errorf("%w", err)
	}

	for i := range commits {
		if commits[i].Committer.ID == dependabotID {
			tools = append(tools, checker.Tool{
				Name:  "Dependabot",
				URL:   asPointer("https://github.com/dependabot"),
				Desc:  asPointer("Automated dependency updates built into GitHub"),
				Files: []checker.File{{}},
			})
			break
		}
	}

	return checker.DependencyUpdateToolData{Tools: tools}, nil
}

var checkDependencyFileExists fileparser.DoWhileTrueOnFilename = func(name string, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("checkDependencyFileExists requires exactly one argument: %w", errInvalidArgLength)
	}
	ptools, ok := args[0].(*[]checker.Tool)
	if !ok {
		return false, fmt.Errorf(
			"checkDependencyFileExists requires an argument of type: *[]checker.Tool: %w", errInvalidArgType)
	}

	switch strings.ToLower(name) {
	case ".github/dependabot.yml", ".github/dependabot.yaml":
		*ptools = append(*ptools, checker.Tool{
			Name: "Dependabot",
			URL:  asPointer("https://github.com/dependabot"),
			Desc: asPointer("Automated dependency updates built into GitHub"),
			Files: []checker.File{
				{
					Path:   name,
					Type:   finding.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
			},
		})

		// https://docs.renovatebot.com/configuration-options/
	case ".github/renovate.json", ".github/renovate.json5", ".renovaterc.json", "renovate.json",
		"renovate.json5", ".renovaterc":
		*ptools = append(*ptools, checker.Tool{
			Name: "RenovateBot",
			URL:  asPointer("https://github.com/renovatebot/renovate"),
			Desc: asPointer("Automated dependency updates. Multi-platform and multi-language."),
			Files: []checker.File{
				{
					Path:   name,
					Type:   finding.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
			},
		})
	case ".pyup.yml":
		*ptools = append(*ptools, checker.Tool{
			Name: "PyUp",
			URL:  asPointer("https://pyup.io/"),
			Desc: asPointer("Automated dependency updates for Python."),
			Files: []checker.File{
				{
					Path:   name,
					Type:   finding.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
			},
		})
	}

	// Continue iterating, even if we have found a tool.
	// It's needed for all probes results to be populated.
	return true, nil
}

func asPointer(s string) *string {
	return &s
}

func asBoolPointer(b bool) *bool {
	return &b
}
