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

package raw

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

var (
	remediationBranch   string
	remediationRepo     string
	remediationOnce     *sync.Once
	remediationSetupErr error
)

var (
	workflowText = "update your workflow using https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s"
	//nolint
	workflowMarkdown = "update your workflow using [https://app.stepsecurity.io](https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s)"
)

//nolint:gochecknoinits
func init() {
	remediationOnce = new(sync.Once)
}

func remdiationSetup(c *checker.CheckRequest) error {
	remediationOnce.Do(func() {
		// Get the branch for remediation.
		b, err := c.RepoClient.GetDefaultBranch()
		if err != nil && !errors.Is(err, clients.ErrUnsupportedFeature) {
			remediationSetupErr = err
			return
		}
		if b.Name != nil {
			remediationBranch = *b.Name
			uri := c.Repo.URI()
			parts := strings.Split(uri, "/")
			if len(parts) != 3 {
				remediationSetupErr = fmt.Errorf("%w: %s", errInvalidArgLength, uri)
				return
			}
			remediationRepo = fmt.Sprintf("%s/%s", parts[1], parts[2])
		}
	})

	return remediationSetupErr
}

func createWorkflowPermissionRemediation(filepath string) *checker.Remediation {
	return createWorkflowRemediation(filepath, "permissions")
}

func createWorkflowPinningRemediation(filepath string) *checker.Remediation {
	return createWorkflowRemediation(filepath, "pin")
}

func createWorkflowRemediation(path, t string) *checker.Remediation {
	p := strings.TrimPrefix(path, ".github/workflows/")
	if remediationBranch == "" || remediationRepo == "" {
		return nil
	}

	text := fmt.Sprintf(workflowText, remediationRepo, p, remediationBranch, t)
	markdown := fmt.Sprintf(workflowMarkdown, remediationRepo, p, remediationBranch, t)

	return &checker.Remediation{
		HelpText:     text,
		HelpMarkdown: markdown,
	}
}
