// Copyright 2022 Security Scorecard Authors
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

package remediation

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

var (
	branch   string
	repo     string
	once     *sync.Once
	setupErr error
)

var errInvalidArg = errors.New("invalid argument")

var (
	workflowText = "update your workflow using https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s"
	//nolint
	workflowMarkdown = "update your workflow using [https://app.stepsecurity.io](https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s)"
)

//nolint:gochecknoinits
func init() {
	once = new(sync.Once)
}

// Setup sets up remediation code.
func Setup(c *checker.CheckRequest) error {
	once.Do(func() {
		// Get the branch for remediation.
		b, err := c.RepoClient.GetDefaultBranch()
		if err != nil {
			if !errors.Is(err, clients.ErrUnsupportedFeature) {
				setupErr = err
			}
			return
		}

		if b != nil && b.Name != nil {
			branch = *b.Name
			uri := c.RepoClient.URI()
			parts := strings.Split(uri, "/")
			if len(parts) != 3 {
				setupErr = fmt.Errorf("%w: enpty: %s", errInvalidArg, uri)
				return
			}
			repo = fmt.Sprintf("%s/%s", parts[1], parts[2])
		}
	})
	return setupErr
}

// CreateWorkflowPermissionRemediation create remediation for workflow permissions.
func CreateWorkflowPermissionRemediation(filepath string) *checker.Remediation {
	return createWorkflowRemediation(filepath, "permissions")
}

// CreateWorkflowPinningRemediation create remediaiton for pinninn GH Actions.
func CreateWorkflowPinningRemediation(filepath string) *checker.Remediation {
	return createWorkflowRemediation(filepath, "pin")
}

func createWorkflowRemediation(path, t string) *checker.Remediation {
	p := strings.TrimPrefix(path, ".github/workflows/")
	if branch == "" || repo == "" {
		return nil
	}

	text := fmt.Sprintf(workflowText, repo, p, branch, t)
	markdown := fmt.Sprintf(workflowMarkdown, repo, p, branch, t)

	return &checker.Remediation{
		HelpText:     text,
		HelpMarkdown: markdown,
	}
}
