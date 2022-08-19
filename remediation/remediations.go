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
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"

	"github.com/ossf/scorecard/v4/checker"
)

var (
	workflowText = "update your workflow using https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s"
	//nolint
	workflowMarkdown  = "update your workflow using [https://app.stepsecurity.io](https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s)"
	dockerfilePinText = "pin your Docker image by updating %[1]s to %[1]s@%s"
)

// CreateWorkflowPermissionRemediation create remediation for workflow permissions.
func CreateWorkflowPermissionRemediation(r checker.RemediationMetadata, filepath string) *checker.Remediation {
	return createWorkflowRemediation(r, filepath, "permissions")
}

// CreateWorkflowPinningRemediation create remediaiton for pinninn GH Actions.
func CreateWorkflowPinningRemediation(r checker.RemediationMetadata, filepath string) *checker.Remediation {
	return createWorkflowRemediation(r, filepath, "pin")
}

func createWorkflowRemediation(r checker.RemediationMetadata, path, t string) *checker.Remediation {
	p := strings.TrimPrefix(path, ".github/workflows/")
	if r.Branch == "" || r.Repo == "" {
		return nil
	}

	text := fmt.Sprintf(workflowText, r.Repo, p, r.Branch, t)
	markdown := fmt.Sprintf(workflowMarkdown, r.Repo, p, r.Branch, t)

	return &checker.Remediation{
		HelpText:     text,
		HelpMarkdown: markdown,
	}
}

// CreateDockerfilePinningRemediation create remediaiton for pinning Dockerfile images.
func CreateDockerfilePinningRemediation(r checker.RemediationMetadata, name *string) *checker.Remediation {
	if name == nil {
		return nil
	}
	hash, err := crane.Digest(*name)
	if err != nil {
		return nil
	}

	text := fmt.Sprintf(dockerfilePinText, *name, hash)
	markdown := text

	return &checker.Remediation{
		HelpText:     text,
		HelpMarkdown: markdown,
	}
}
