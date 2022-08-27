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
	info []checker.SecurityPolicyInformation
	uri  string
	file checker.File
}

// SecurityPolicy checks for presence of security policy.
func SecurityPolicy(c *checker.CheckRequest) (checker.SecurityPolicyData, error) {
	data := securityPolicyFilesWithURI{
		uri:  "",
		info: make([]checker.SecurityPolicyInformation, 0),
		file: checker.File{
			Path:      "",
			Snippet:   "",
			Offset:    0,
			EndOffset: 0,
			Type:      checker.FileTypeNone,
		},
	}
	err := fileparser.OnAllFilesDo(c.RepoClient, isSecurityPolicyFile, &data)
	if err != nil {
		return checker.SecurityPolicyData{}, err
	}
	// If we found files in the repo, return immediately.
	if data.file.Type != checker.FileTypeNone {
		err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
			Pattern:       data.file.Path,
			CaseSensitive: false,
		}, checkSecurityPolicyFileContent, &data.file, &data.info)
		if err != nil {
			return checker.SecurityPolicyData{}, err
		}
		return checker.SecurityPolicyData{File: data.file, Information: data.info}, nil
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
	if data.file.Type != checker.FileTypeNone {
		filePattern := data.file.Path
		// undo path.Join in isSecurityPolicyFile
		if data.file.Type == checker.FileTypeURL {
			filePattern = strings.Replace(data.file.Path, data.uri+"/", "", 1)
		}
		err := fileparser.OnMatchingFileContentDo(dotGitHubClient, fileparser.PathMatcher{
			Pattern:       filePattern,
			CaseSensitive: false,
		}, checkSecurityPolicyFileContent, &data.file, &data.info)
		if err != nil {
			return checker.SecurityPolicyData{}, err
		}
	}
	return checker.SecurityPolicyData{File: data.file, Information: data.info}, nil
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
		tempType := checker.FileTypeText
		if pdata.uri != "" {
			// TODO: is joining even needed?
			tempPath = path.Join(pdata.uri, tempPath)
			// FileTypeURL is used in Security-Policy to
			// only denote for the details report that the
			// policy was found at the org level rather
			// than the repo level
			tempType = checker.FileTypeURL
		}
		pdata.file = checker.File{
			Path:   tempPath,
			Type:   tempType,
			Offset: checker.OffsetDefault,
		}
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
	if len(args) != 2 {
		return false, fmt.Errorf(
			"checkSecurityPolicyFileContent requires exactly two arguments: %w", errInvalidArgLength)
	}
	pfiles, ok := args[0].(*checker.File)
	if !ok {
		return false, fmt.Errorf(
			"checkSecurityPolicyFileContent requires argument of type *checker.File: %w", errInvalidArgType)
	}
	pinfo, ok := args[1].(*[]checker.SecurityPolicyInformation)
	if !ok {
		return false, fmt.Errorf(
			"%s requires argument of type *[]checker.SecurityPolicyInformation: %w",
			"checkSecurityPolicyFileContent", errInvalidArgType)
	}

	if len(content) == 0 {
		// perhaps there are more policy files somewhere else,
		// keep looking (true)
		return true, nil
	}

	if pfiles != nil && (*pinfo) != nil {
		// preserve file type
		tempType := pfiles.Type
		urlsL, emailsL, discvulsL := collectPolicyHits(string(content))

		*pfiles = checker.File{
			Path:   path,
			Type:   tempType,
			Offset: checker.OffsetDefault,
			// convey the length/amount of content using
			// the EndOffset as the len to EOF used in eval
			EndOffset: uint(len(content)),
		}

		if len(urlsL) > 0 {
			(*pinfo) = append((*pinfo), checker.SecurityPolicyInformation{
				InformationType:  checker.SecurityPolicyInformationTypeLink,
				InformationValue: urlsL,
			})
		}

		if len(emailsL) > 0 {
			(*pinfo) = append((*pinfo), checker.SecurityPolicyInformation{
				InformationType:  checker.SecurityPolicyInformationTypeEmail,
				InformationValue: emailsL,
			})
		}

		if len(discvulsL) > 0 {
			(*pinfo) = append((*pinfo), checker.SecurityPolicyInformation{
				InformationType:  checker.SecurityPolicyInformationTypeText,
				InformationValue: discvulsL,
			})
		}
	} else {
		e := sce.WithMessage(sce.ErrScorecardInternal, "bad file or information reference")
		return false, e
	}

	// stop here found something, no need to look further (false)
	return false, nil
}

func collectPolicyHits(policyContent string) ([]string, []string, []string) {
	// pattern for URLs
	reURL := regexp.MustCompile(`(http|https)://[a-zA-Z0-9./?=_%:-]*`)
	// pattern for emails
	reEML := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,6}\b`)
	// pattern for 1 to 4 digit numbers
	// or
	// strings 'disclos' as in "disclosure" or 'vuln' as in "vulnerability"
	reDIG := regexp.MustCompile(`(?i)(\b*[0-9]{1,4}\b|(Disclos|Vuln))`)

	urlUniqList := uniqueInPolicy(reURL.FindAllString(policyContent, -1))

	emailUniqList := uniqueInPolicy(reEML.FindAllString(policyContent, -1))

	// not really looking unique sets of numbers or words
	discvulsList := reDIG.FindAllString(policyContent, -1)

	return urlUniqList, emailUniqList, discvulsList
}

//
// Returns unique items in a slice
// src inspiration: https://codereview.stackexchange.com/questions/191238/return-unique-items-in-a-go-slice
//
func uniqueInPolicy(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
