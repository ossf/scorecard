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
package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v3/checks"
	docs "github.com/ossf/scorecard/v3/docs/checks"
)

var (
	allowedRisks     = map[string]bool{"Critical": true, "High": true, "Medium": true, "Low": true}
	allowedRepoTypes = map[string]bool{"GitHub": true, "local": true}
	supportedAPIs    = map[string][]string{
		// InitRepo is supported for local reepos in general. However, in the context of checks,
		// this is only used to look up remote data, e.g. in Fuzzinng check.
		// So we only have "GitHub" supported.
		"InitRepo":                   {"GitHub"},
		"URI":                        {"GitHub", "local"},
		"IsArchived":                 {"GitHub"},
		"ListFiles":                  {"GitHub", "local"},
		"GetFileContent":             {"GitHub", "local"},
		"ListMergedPRs":              {"GitHub"},
		"ListBranches":               {"GitHub"},
		"GetDefaultBranch":           {"GitHub"},
		"ListCommits":                {"GitHub"},
		"ListReleases":               {"GitHub"},
		"ListContributors":           {"GitHub"},
		"ListSuccessfulWorkflowRuns": {"GitHub"},
		"ListCheckRunsForRef":        {"GitHub"},
		"ListStatuses":               {"GitHub"},
		"Search":                     {"GitHub", "local"},
		"Close":                      {"GitHub", "local"},
	}
)

func listCheckFiles() (map[string]string, error) {
	checkFiles := make(map[string]string)
	// Use regex to determine the file that contains the entry point.
	regex := regexp.MustCompile(`const\s+[^"]*=\s+"(.*)"`)
	files, err := ioutil.ReadDir("checks/")
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadDir: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".go") || file.IsDir() {
			continue
		}

		fullpath := path.Join("checks/", file.Name())
		content, err := ioutil.ReadFile(fullpath)
		if err != nil {
			return nil, fmt.Errorf("ioutil.ReadFile: %s: %w", fullpath, err)
		}

		res := regex.FindStringSubmatch(string(content))
		if len(res) != 2 {
			continue
		}

		r := res[1]
		if entry, exists := checkFiles[r]; exists {
			//nolint:goerr113
			return nil, fmt.Errorf("check %s already exists: %v", r, entry)
		}
		checkFiles[r] = fullpath
	}
	return checkFiles, nil
}

func extractAPINames() ([]string, error) {
	fns := []string{}
	interfaceRe := regexp.MustCompile(`type\s+RepoClient\s+interface\s+{\s*`)
	content, err := ioutil.ReadFile("clients/repo_client.go")
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadFile: %s: %w", "clients/repo_client.go", err)
	}

	locs := interfaceRe.FindIndex(content)
	if len(locs) != 2 || locs[1] == 0 {
		//nolint:goerr113
		return nil, fmt.Errorf("FindIndex: cannot find Doc interface definition")
	}

	nameRe := regexp.MustCompile(`[\s]+([A-Z]\S+)\s*\(.*\).+[\n]+`)
	matches := nameRe.FindAllStringSubmatch(string(content[locs[1]-1:]), -1)
	if len(matches) == 0 {
		//nolint:goerr113
		return nil, fmt.Errorf("FindAllStringSubmatch: no match found")
	}

	for _, v := range matches {
		if len(v) != 2 {
			//nolint:goerr113
			return nil, fmt.Errorf("invalid length: %d", len(v))
		}
		fns = append(fns, v[1])
	}
	return fns, nil
}

func isSubsetOf(a, b []string) bool {
	mb := make(map[string]bool)
	for _, vb := range b {
		mb[vb] = true
	}
	for _, va := range a {
		if _, exists := mb[va]; !exists {
			return false
		}
	}
	return true
}

func contains(l []string, elt string) bool {
	for _, v := range l {
		if v == elt {
			return true
		}
	}
	return false
}

func supportedInterfacesFromImplementation(checkName string, checkFiles map[string]string) ([]string, error) {
	// Special case. No APIs are used,
	// but we need the repo name for a db lookup.
	if checkName == "CII-Best-Practices" {
		return []string{"GitHub"}, nil
	}

	// Create our map.
	s := make(map[string]bool)
	for k := range allowedRepoTypes {
		s[k] = true
	}

	// Read the source file for the check.
	pathfn, exists := checkFiles[checkName]
	if !exists {
		//nolint:goerr113
		return nil, fmt.Errorf("check %s does not exists", checkName)
	}

	content, err := ioutil.ReadFile(pathfn)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadFile: %s: %w", pathfn, err)
	}

	// For each API, check if it's used or not.
	// Adjust supported repo types accordingly.
	for api, supportedInterfaces := range supportedAPIs {
		regex := fmt.Sprintf(`\.%s`, api)
		re := regexp.MustCompile(regex)
		r := re.Match(content)
		if r {
			for k := range allowedRepoTypes {
				if !contains(supportedInterfaces, k) {
					delete(s, k)
				}
			}
		}
	}

	r := []string{}
	for k := range s {
		r = append(r, k)
	}
	return r, nil
}

func validateRepoTypeAPIs(checkName string, repoTypes []string, checkFiles map[string]string) error {
	// For now, we only list APIs in a check's implementation.
	// Long-term, we should use the callgraph feature using
	// https://github.com/golang/tools/blob/master/cmd/callgraph/main.go
	l, err := supportedInterfacesFromImplementation(checkName, checkFiles)
	if err != nil {
		return fmt.Errorf("supportedInterfacesFromImplementation: %w", err)
	}

	if !cmp.Equal(l, repoTypes, cmpopts.SortSlices(func(x, y string) bool { return x < y })) {
		//nolint: goerr113
		return fmt.Errorf("%s: got diff: %s", checkName, cmp.Diff(l, repoTypes))
	}
	return nil
}

func validateAPINames() error {
	// Extract API names.
	fns, err := extractAPINames()
	if err != nil {
		return fmt.Errorf("invalid functions: %w", err)
	}

	// Validate function names.
	functions := []string{}
	for v := range supportedAPIs {
		functions = append(functions, v)
	}

	if !cmp.Equal(functions, fns, cmpopts.SortSlices(func(x, y string) bool { return x < y })) {
		//nolint:goerr113
		return fmt.Errorf("got diff: %s", cmp.Diff(functions, fns))
	}

	return nil
}

func main() {
	m, err := docs.Read()
	if err != nil {
		panic(fmt.Sprintf("docs.Read: %v", err))
	}

	if err := validateAPINames(); err != nil {
		panic(fmt.Sprintf("cannot extract function names: %v", err))
	}

	checkFiles, err := listCheckFiles()
	if err != nil {
		panic(err)
	}

	allChecks := checks.AllChecks
	for check := range allChecks {
		c, e := m.GetCheck(check)
		if e != nil {
			panic(fmt.Sprintf("GetCheck: %v: %s", e, check))
		}

		if check != c.GetName() {
			panic(fmt.Sprintf("invalid checkName: %s != %s", check, c.GetName()))
		}
		if c.GetDescription() == "" {
			panic(fmt.Sprintf("description for checkName: %s is empty", check))
		}
		if strings.TrimSpace(strings.Join(c.GetRemediation(), "")) == "" {
			panic(fmt.Sprintf("remediation for checkName: %s is empty", check))
		}
		if c.GetShort() == "" {
			panic(fmt.Sprintf("short for checkName: %s is empty", check))
		}
		if len(c.GetTags()) == 0 {
			panic(fmt.Sprintf("tags for checkName: %s is empty", check))
		}
		r := c.GetRisk()
		if _, exists := allowedRisks[r]; !exists {
			panic(fmt.Sprintf("risk for checkName: %s is invalid: '%s'", check, r))
		}
		repoTypes := c.GetSupportedRepoTypes()
		if len(repoTypes) == 0 {
			panic(fmt.Sprintf("repos for checkName: %s is empty", check))
		}
		for _, rt := range repoTypes {
			if _, exists := allowedRepoTypes[rt]; !exists {
				panic(fmt.Sprintf("repo type for checkName: %s is invalid: '%s'", check, rt))
			}
		}

		// Validate that the check only calls API the interface supports.
		if err := validateRepoTypeAPIs(check, repoTypes, checkFiles); err != nil {
			panic(fmt.Sprintf("validateRepoTypeAPIs: %v", err))
		}
	}
	for _, check := range m.GetChecks() {
		if _, exists := allChecks[check.GetName()]; !exists {
			panic(fmt.Sprintf("check present in checks.yaml is not part of `checks.AllChecks`: %s", check.GetName()))
		}
	}
}
