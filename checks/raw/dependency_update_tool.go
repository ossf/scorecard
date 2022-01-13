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
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
)

// DependencyUpdateTool is the exported name for Depdendency-Update-Tool.
func DependencyUpdateTool(c clients.RepoClient) (checker.DependencyUpdateToolData, error) {
	var tools []checker.Tool
	err := fileparser.CheckIfFileExistsV6(c, checkDependencyFileExists, &tools)
	if err != nil {
		return checker.DependencyUpdateToolData{}, fmt.Errorf("%w", err)
	}

	// No error, return the tools.
	return checker.DependencyUpdateToolData{Tools: tools}, nil
}

func checkDependencyFileExists(name string, data fileparser.FileCbData) (bool, error) {
	ptools, ok := data.(*[]checker.Tool)
	if !ok {
		// This never happens.
		panic("invalid type")
	}

	switch strings.ToLower(name) {
	case ".github/dependabot.yml":
		*ptools = append(*ptools, checker.Tool{
			Name: "Dependabot",
			URL:  "https://github.com/dependabot",
			Desc: "Automated dependency updates built into GitHub",
			ConfigFiles: []checker.File{
				{
					Path:   name,
					Type:   checker.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
			},
		})

		// https://docs.renovatebot.com/configuration-options/
	case ".github/renovate.json", ".github/renovate.json5", ".renovaterc.json", "renovate.json",
		"renovate.json5", ".renovaterc":
		*ptools = append(*ptools, checker.Tool{
			Name: "Renovabot",
			URL:  "https://github.com/renovatebot/renovate",
			Desc: "Automated dependency updates. Multi-platform and multi-language.",
			ConfigFiles: []checker.File{
				{
					Path:   name,
					Type:   checker.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
			},
		})
	default:
		// Continue iterating.
		return true, nil
	}

	// We found a file, no need to continue iterating.
	return false, nil
}
