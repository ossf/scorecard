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
	"strings"

	"go.uber.org/zap"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/checks/fileparser"
	"github.com/ossf/scorecard/v3/clients/githubrepo"
	sce "github.com/ossf/scorecard/v3/errors"
)

// SecurityPolicyData contains the raw results.
type SecurityPolicyData struct {
	// Files contains a list of files.
	Files []File
}

// SecurityPolicy checks for presence of security policy.
func SecurityPolicy(c *checker.CheckRequest) (SecurityPolicyData, error) {
	// TODO: not supported for local clients.
	var rawData SecurityPolicyData

	// Check repository for repository-specific policy.
	// https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file.
	onFile := func(name string, dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
		rawData, ok := data.(*SecurityPolicyData)
		if !ok {
			// This never happens.
			panic("invalid type")
		}
		if strings.EqualFold(name, "security.md") ||
			strings.EqualFold(name, ".github/security.md") ||
			strings.EqualFold(name, "docs/security.md") {
			rawData.Files = append(rawData.Files, File{
				Path:   name,
				Type:   checker.FileTypeSource,
				Offset: checker.OffsetDefault,
			})
			return false, nil
		} else if isSecurityRstFound(name) {
			rawData.Files = append(rawData.Files, File{
				Path:   name,
				Type:   checker.FileTypeSource,
				Offset: checker.OffsetDefault,
			})
			return false, nil
		}
		return true, nil
	}

	err := fileparser.CheckIfFileExists(c, onFile, &rawData)
	if err != nil {
		return SecurityPolicyData{}, err
	}

	// If we found files in the repo, return immediately.
	if len(rawData.Files) > 0 {
		return rawData, nil
	}

	// https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file.
	logger, err := githubrepo.NewLogger(zap.InfoLevel)
	if err != nil {
		return SecurityPolicyData{}, fmt.Errorf("%w", err)
	}
	dotGitHub := &checker.CheckRequest{
		Ctx:        c.Ctx,
		Dlogger:    c.Dlogger,
		RepoClient: githubrepo.CreateGithubRepoClient(c.Ctx, logger),
		Repo:       c.Repo.Org(),
	}

	err = dotGitHub.RepoClient.InitRepo(dotGitHub.Repo)
	switch {
	case err == nil:
		defer dotGitHub.RepoClient.Close()
		onFile = func(name string, dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
			rawData, ok := data.(*SecurityPolicyData)
			if !ok {
				// This never happens.
				panic("invalid type")
			}
			if strings.EqualFold(name, "security.md") ||
				strings.EqualFold(name, ".github/security.md") ||
				strings.EqualFold(name, "docs/security.md") {
				rawData.Files = append(rawData.Files, File{
					Path:   name,
					Type:   checker.FileTypeURL,
					Offset: checker.OffsetDefault,
				})
				return false, nil
			}
			return true, nil
		}
		err = fileparser.CheckIfFileExists(dotGitHub, onFile, &rawData)
		if err != nil {
			return SecurityPolicyData{}, err
		}

	case errors.Is(err, sce.ErrRepoUnreachable):
		break
	default:
		return SecurityPolicyData{}, err
	}

	return rawData, err
}

func isSecurityRstFound(name string) bool {
	if strings.EqualFold(name, "doc/security.rst") {
		return true
	} else if strings.EqualFold(name, "docs/security.rst") {
		return true
	}
	return false
}
