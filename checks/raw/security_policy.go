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
	if len(data.files) > 0 {
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
		tempType := checker.FileTypeSource
		if pdata.uri != "" {
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
		strings.EqualFold(name, "doc/security.rst") ||
		strings.EqualFold(name, "docs/security.rst")
}
