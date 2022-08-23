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
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/log"
)

type securityPolicyFilesWithURI struct {
	uri   string
	files []checker.File
}

// SecurityPolicy checks for presence of security policy.
func SecurityPolicy(c *checker.CheckRequest) (checker.SecurityPolicyData, error) {
	data := securityPolicyFilesWithURI{
		uri:   "",
		files: make([]checker.File, 0),
	}
	err := fileparser.OnAllFilesDo(c.RepoClient, isSecurityPolicyFile, &data)
	if err != nil {
		return checker.SecurityPolicyData{}, err
	}
	// If we found files in the repo, return immediately.
	// TODO: // assert that at this point only 1 file is returned by isSecurityPolicyFile
	if len(data.files) > 0 {
		err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
			Pattern:       data.files[0].Path,
			CaseSensitive: false,
		}, checkSecurityPolicyFileContent, &data.files)
		if err != nil {
			return checker.SecurityPolicyData{}, err
		}
		return checker.SecurityPolicyData{Files: data.files}, nil
	}

	// Check if present in parent org.
	// https#://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file.
	// TODO(1491): Make this non-GitHub specific.
	logger := log.NewLogger(log.InfoLevel)
	dotGitHubClient := githubrepo.CreateGithubRepoClient(c.Ctx, logger)
	err = dotGitHubClient.InitRepo(c.Repo.Org(), clients.HeadSHA)
	switch {
	case err == nil:
		defer dotGitHubClient.Close()
		data.uri = dotGitHubClient.URI()
		err = fileparser.OnAllFilesDo(dotGitHubClient, isSecurityPolicyFile, &data)
		if err != nil {
			return checker.SecurityPolicyData{}, err
		}

	case errors.Is(err, sce.ErrRepoUnreachable):
		break
	default:
		return checker.SecurityPolicyData{}, err
	}

	// Return raw results.
	if len(data.files) > 0 {
		filePattern := data.files[0].Path
		// undo path.Join in isSecurityPolicyFile
		if data.files[0].Type == checker.FileTypeURL {
			filePattern = strings.Replace(data.files[0].Path, data.uri+"/", "", 1)
		}
		err := fileparser.OnMatchingFileContentDo(dotGitHubClient, fileparser.PathMatcher{
			Pattern:       filePattern,
			CaseSensitive: false,
		}, checkSecurityPolicyFileContent, &data.files)
		if err != nil {
			return checker.SecurityPolicyData{}, err
		}
	}
	return checker.SecurityPolicyData{Files: data.files}, nil
}

// Check repository for repository-specific policy.
// https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file.
var isSecurityPolicyFile fileparser.DoWhileTrueOnFilename = func(name string, args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("isSecurityPolicyFile requires exactly one argument: %w", errInvalidArgLength)
	}
	pdata, ok := args[0].(*securityPolicyFilesWithURI)
	if !ok {
		return false, fmt.Errorf("invalid arg type: %w", errInvalidArgType)
	}
	if isSecurityPolicyFilename(name) {
		tempPath := name
		// TODO: really should be FileTypeText (.md, .adoc, etc)
		tempType := checker.FileTypeSource
		if pdata.uri != "" {
			// TODO: is joining even needed?
			tempPath = path.Join(pdata.uri, tempPath)
			tempType = checker.FileTypeURL
		}
		pdata.files = append(pdata.files, checker.File{
			Path:   tempPath,
			Type:   tempType,
			Offset: checker.OffsetDefault,
		})
		return false, nil
	}
	return true, nil
}

func isSecurityPolicyFilename(name string) bool {
	return strings.EqualFold(name, "security.md") ||
		strings.EqualFold(name, ".github/security.md") ||
		strings.EqualFold(name, "docs/security.md") ||
		strings.EqualFold(name, "security.adoc") ||
		strings.EqualFold(name, ".github/security.adoc") ||
		strings.EqualFold(name, "docs/security.adoc") ||
		strings.EqualFold(name, "security.rst") ||
		strings.EqualFold(name, ".github/security.rst") ||
		strings.EqualFold(name, "doc/security.rst") ||
		strings.EqualFold(name, "docs/security.rst")
}

var checkSecurityPolicyFileContent fileparser.DoWhileTrueOnFileContent = func(path string, content []byte,
	args ...interface{},
) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf(
			"checkSecurityPolicyFileContent requires exactly one argument: %w", errInvalidArgLength)
	}
	pfiles, ok := args[0].(*[]checker.File)
	if !ok {
		return false, fmt.Errorf(
			"checkSecurityPolicyFileContent requires argument of type *[]checker.File: %w", errInvalidArgType)
	}

	if len(content) == 0 {
		// perhaps there are more policy files somewhere else,
		// keep looking (true)
		return true, nil
	}

	// TODO: there is an assertion here that if content has length
	//       then (*pfiles)[0] is not nil (can that be checked)
	//       (so is this test necessary?)
	if (*pfiles) != nil {
		// preserve file type
		tempType := (*pfiles)[0].Type
		urls, emails, discvuls := countUniquePolicyHits(string(content))
		contentMetrics := fmt.Sprintf("%d,%d,%d,%d", len(content), urls, emails, discvuls)
		(*pfiles)[0] = checker.File{
			Path:    path,
			Type:    tempType,
			Offset:  checker.OffsetDefault,
			Snippet: contentMetrics,
		}
	}

	// stop here found something, no need to look further (false)
	return false, nil
}

func countUniquePolicyHits(policyContent string) (int, int, int) {
	var urls, emails, discvuls int

	// pattern for URLs
	reURL := regexp.MustCompile(`(http|https)://[a-zA-Z0-9./?=_%:-]*`)
	// pattern for emails
	reEML := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,6}\b`)
	// pattern for 1 to 4 digit numbers
	// or
	// strings 'disclos' as in "disclosure" or 'vuln' as in "vulnerability"
	reDIG := regexp.MustCompile(`(?i)(\b*[0-9]{1,4}\b|(Disclos|Vuln))`)

	rURL := reURL.FindAllString(policyContent, -1)
	rEML := reEML.FindAllString(policyContent, -1)
	rDIG := reDIG.FindAllString(policyContent, -1)

	urls = countUniqueInPolicy(rURL)

	emails = countUniqueInPolicy(rEML)

	// not really looking unique sets of numbers or words
	// and take the raw count of hits
	discvuls = len(rDIG)

	return urls, emails, discvuls
}

//
// Returns a count of unique items in a slice
// src inspiration: https://codereview.stackexchange.com/questions/191238/return-unique-items-in-a-go-slice
//
func countUniqueInPolicy(strSlice []string) int {
	count := 0
	keys := make(map[string]bool)
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			count += 1
		}
	}
	return count
}
