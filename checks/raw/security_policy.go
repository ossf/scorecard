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
	"bufio"
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
	files []checker.SecurityPolicyFile
}

// SecurityPolicy checks for presence of security policy
// and applicable content discovered by checkSecurityPolicyFileContent().
func SecurityPolicy(c *checker.CheckRequest) (checker.SecurityPolicyData, error) {
	data := securityPolicyFilesWithURI{
		uri: "", files: make([]checker.SecurityPolicyFile, 0),
	}
	err := fileparser.OnAllFilesDo(c.RepoClient, isSecurityPolicyFile, &data)
	if err != nil {
		return checker.SecurityPolicyData{}, err
	}
	// If we found files in the repo, return immediately.
	if len(data.files) > 0 {
		for idx := range data.files {
			err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
				Pattern:       data.files[idx].File.Path,
				CaseSensitive: false,
			}, checkSecurityPolicyFileContent, &data.files[idx].File, &data.files[idx].Information)
			if err != nil {
				return checker.SecurityPolicyData{}, err
			}
		}
		return checker.SecurityPolicyData{PolicyFiles: data.files}, nil
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
		for idx := range data.files {
			filePattern := data.files[idx].File.Path
			// undo path.Join in isSecurityPolicyFile just
			// for this call to OnMatchingFileContentsDo
			if data.files[idx].File.Type == checker.FileTypeURL {
				filePattern = strings.Replace(filePattern, data.uri+"/", "", 1)
			}
			err := fileparser.OnMatchingFileContentDo(dotGitHubClient, fileparser.PathMatcher{
				Pattern:       filePattern,
				CaseSensitive: false,
			}, checkSecurityPolicyFileContent, &data.files[idx].File, &data.files[idx].Information)
			if err != nil {
				return checker.SecurityPolicyData{}, err
			}
		}
	}
	return checker.SecurityPolicyData{PolicyFiles: data.files}, nil
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
			// report complete path for org-based policy files
			tempPath = path.Join(pdata.uri, tempPath)
			// FileTypeURL is used in Security-Policy to
			// only denote for the details report that the
			// policy was found at the org level rather
			// than the repo level
			tempType = checker.FileTypeURL
		}
		pdata.files = append(pdata.files, checker.SecurityPolicyFile{
			File: checker.File{
				Path:     tempPath,
				Type:     tempType,
				Offset:   checker.OffsetDefault,
				FileSize: checker.OffsetDefault,
			},
			Information: make([]checker.SecurityPolicyInformation, 0),
		})
		// TODO: change 'false' to 'true' when multiple security policy files are supported
		// otherwise this check stops at the first security policy found
		return false, nil
	}
	return true, nil
}

func isSecurityPolicyFilename(name string) bool {
	return strings.EqualFold(name, "security.md") ||
		strings.EqualFold(name, ".github/security.md") ||
		strings.EqualFold(name, "docs/security.md") ||
		strings.EqualFold(name, "security.markdown") ||
		strings.EqualFold(name, ".github/security.markdown") ||
		strings.EqualFold(name, "docs/security.markdown") ||
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
		pfiles.Offset = checker.OffsetDefault
		pfiles.FileSize = uint(len(content))
		policyHits := collectPolicyHits(content)
		if len(policyHits) > 0 {
			(*pinfo) = append((*pinfo), policyHits...)
		}
	} else {
		e := sce.WithMessage(sce.ErrScorecardInternal, "bad file or information reference")
		return false, e
	}

	// stop here found something, no need to look further (false)
	return false, nil
}

func collectPolicyHits(policyContent []byte) []checker.SecurityPolicyInformation {
	var hits []checker.SecurityPolicyInformation

	// pattern for URLs
	reURL := regexp.MustCompile(`(http|https)://[a-zA-Z0-9./?=_%:-]*`)
	// pattern for emails
	reEML := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,6}\b`)
	// pattern for 1 to 4 digit numbers
	// or
	// strings 'disclos' as in "disclosure" or 'vuln' as in "vulnerability"
	reDIG := regexp.MustCompile(`(?i)(\b*[0-9]{1,4}\b|(Disclos|Vuln))`)

	lineNum := 0
	for {
		advance, token, err := bufio.ScanLines(policyContent, true)
		if advance == 0 || err != nil {
			break
		}

		lineNum += 1
		if len(token) != 0 {
			for _, indexes := range reURL.FindAllIndex(token, -1) {
				hits = append(hits, checker.SecurityPolicyInformation{
					InformationType: checker.SecurityPolicyInformationTypeLink,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      string(token[indexes[0]:indexes[1]]), // Snippet of match
						LineNumber: uint(lineNum),                        // line number in file
						Offset:     uint(indexes[0]),                     // Offset in the line
					},
				})
			}
			for _, indexes := range reEML.FindAllIndex(token, -1) {
				hits = append(hits, checker.SecurityPolicyInformation{
					InformationType: checker.SecurityPolicyInformationTypeEmail,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      string(token[indexes[0]:indexes[1]]), // Snippet of match
						LineNumber: uint(lineNum),                        // line number in file
						Offset:     uint(indexes[0]),                     // Offset in the line
					},
				})
			}
			for _, indexes := range reDIG.FindAllIndex(token, -1) {
				hits = append(hits, checker.SecurityPolicyInformation{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      string(token[indexes[0]:indexes[1]]), // Snippet of match
						LineNumber: uint(lineNum),                        // line number in file
						Offset:     uint(indexes[0]),                     // Offset in the line
					},
				})
			}
		}
		if advance <= len(policyContent) {
			policyContent = policyContent[advance:]
		}
	}

	return hits
}
