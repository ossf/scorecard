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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
)

const (
	dependabotID = 49699333
)

type dependencyToolConfig struct {
	Name        string
	URL         string
	Description string
	Paths       []string
}

var dependencyToolConfigs = []dependencyToolConfig{
	{
		Name:        "Dependabot",
		URL:         "https://github.com/dependabot",
		Description: "Automated dependency updates built into GitHub",
		Paths: []string{
			".github/dependabot.yml",
			".github/dependabot.yaml",
		},
	},
	{
		Name:        "RenovateBot",
		URL:         "https://github.com/renovatebot/renovate",
		Description: "Automated dependency updates. Multi-platform and multi-language.",
		Paths: []string{
			"renovate.json",
			"renovate.json5",
			".github/renovate.json",
			".github/renovate.json5",
			".gitlab/renovate.json",
			".gitlab/renovate.json5",
			".renovaterc",
			".renovaterc.json",
			".renovaterc.json5",
		},
	},
	{
		Name:        "scala-steward",
		URL:         "https://github.com/scala-steward-org/scala-steward",
		Description: "Works with Maven, Mill, sbt, and Scala CLI.",
		Paths: []string{
			".scala-steward.conf",
			"scala-steward.conf",
			".github/.scala-steward.conf",
			".github/scala-steward.conf",
			".config/.scala-steward.conf",
			".config/scala-steward.conf",
		},
	},
}

// DependencyUpdateTool is the exported name for Dependency-Update-Tool.
func DependencyUpdateTool(c clients.RepoClient) (checker.DependencyUpdateToolData, error) {
	var tools []checker.Tool
	err := fileparser.OnAllFilesDo(c, checkDependencyFileExists, &tools)
	if err != nil {
		return checker.DependencyUpdateToolData{}, fmt.Errorf("%w", err)
	}

	if len(tools) != 0 {
		return checker.DependencyUpdateToolData{Tools: tools}, nil
	}

	tools, err = findDependencyFiles(c)
	if err != nil {
		if !errors.Is(err, clients.ErrUnsupportedFeature) {
			return checker.DependencyUpdateToolData{}, fmt.Errorf("dependency update tool config lookup: %w", err)
		}
	}
	if len(tools) != 0 {
		return checker.DependencyUpdateToolData{Tools: tools}, nil
	}

	commits, err := c.SearchCommits(clients.SearchCommitsOptions{Author: "dependabot[bot]"})
	if err != nil {
		// TODO https://github.com/ossf/scorecard/issues/1709
		// some repo clients (e.g. local) don't currently have the ability to search commits,
		// but some data is better than none.
		if errors.Is(err, clients.ErrUnsupportedFeature) {
			return checker.DependencyUpdateToolData{Tools: tools}, nil
		}
		return checker.DependencyUpdateToolData{}, fmt.Errorf("dependabot commit search: %w", err)
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

func findDependencyFiles(c clients.RepoClient) ([]checker.Tool, error) {
	var tools []checker.Tool
	for _, config := range dependencyToolConfigs {
		for _, path := range config.Paths {
			reader, err := c.GetFileReader(path)
			if err != nil {
				if errors.Is(err, clients.ErrUnsupportedFeature) {
					return nil, err
				}
				if !errors.Is(err, os.ErrNotExist) {
					return nil, err
				}
				continue
			}
			if reader != nil {
				reader.Close()
			}
			tools = append(tools, dependencyTool(config, path))
			break
		}
	}
	return tools, nil
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
		*ptools = append(*ptools, dependencyTool(dependencyToolConfigs[0], name))
	case "renovate.json", "renovate.json5", ".github/renovate.json", ".github/renovate.json5",
		".gitlab/renovate.json", ".gitlab/renovate.json5", ".renovaterc", ".renovaterc.json",
		".renovaterc.json5":
		*ptools = append(*ptools, dependencyTool(dependencyToolConfigs[1], name))
	case ".scala-steward.conf", "scala-steward.conf", ".github/.scala-steward.conf",
		".github/scala-steward.conf", ".config/.scala-steward.conf", ".config/scala-steward.conf":
		*ptools = append(*ptools, dependencyTool(dependencyToolConfigs[2], name))
	}

	// Continue iterating, even if we have found a tool.
	// It's needed for all probes results to be populated.
	return true, nil
}

func dependencyTool(config dependencyToolConfig, path string) checker.Tool {
	return checker.Tool{
		Name: config.Name,
		URL:  asPointer(config.URL),
		Desc: asPointer(config.Description),
		Files: []checker.File{
			{
				Path:   path,
				Type:   finding.FileTypeSource,
				Offset: checker.OffsetDefault,
			},
		},
	}
}

func asPointer(s string) *string {
	return &s
}

func asBoolPointer(b bool) *bool {
	return &b
}
