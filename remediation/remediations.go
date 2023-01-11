// Copyright 2022 OpenSSF Scorecard Authors
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

	"github.com/google/go-containerregistry/pkg/crane"

	"github.com/ossf/scorecard/v4/checker"
)

var errInvalidArg = errors.New("invalid argument")

var (
	workflowText = "update your workflow using https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s"
	//nolint
	workflowMarkdown  = "update your workflow using [https://app.stepsecurity.io](https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s)"
	dockerfilePinText = "pin your Docker image by updating %[1]s to %[1]s@%s"
)

// TODO fix how this info makes it checks/evaluation.
type RemediationMetadata struct {
	branch string
	repo   string
}

// New returns remediation relevant metadata from a CheckRequest.
func New(c *checker.CheckRequest) (RemediationMetadata, error) {
	if c == nil || c.RepoClient == nil {
		return RemediationMetadata{}, nil
	}

	// Get the branch for remediation.
	branch, err := c.RepoClient.GetDefaultBranchName()
	if err != nil {
		return RemediationMetadata{}, fmt.Errorf("GetDefaultBranchName: %w", err)
	}

	uri := c.RepoClient.URI()
	parts := strings.Split(uri, "/")
	if len(parts) != 3 {
		return RemediationMetadata{}, fmt.Errorf("%w: empty: %s", errInvalidArg, uri)
	}
	repo := fmt.Sprintf("%s/%s", parts[1], parts[2])
	return RemediationMetadata{branch: branch, repo: repo}, nil
}

// CreateWorkflowPermissionRemediation create remediation for workflow permissions.
func (r *RemediationMetadata) CreateWorkflowPermissionRemediation(filepath string) *checker.Remediation {
	return r.createWorkflowRemediation(filepath, "permissions")
}

// CreateWorkflowPinningRemediation create remediaiton for pinninn GH Actions.
func (r *RemediationMetadata) CreateWorkflowPinningRemediation(filepath string) *checker.Remediation {
	return r.createWorkflowRemediation(filepath, "pin")
}

func (r *RemediationMetadata) createWorkflowRemediation(path, t string) *checker.Remediation {
	p := strings.TrimPrefix(path, ".github/workflows/")
	if r.branch == "" || r.repo == "" {
		return nil
	}

	text := fmt.Sprintf(workflowText, r.repo, p, r.branch, t)
	markdown := fmt.Sprintf(workflowMarkdown, r.repo, p, r.branch, t)

	return &checker.Remediation{
		HelpText:     text,
		HelpMarkdown: markdown,
	}
}

func dockerImageName(d *checker.Dependency) (name string, ok bool) {
	if d.Name == nil || *d.Name == "" {
		return "", false
	}
	if d.PinnedAt != nil && *d.PinnedAt != "" {
		return fmt.Sprintf("%s:%s", *d.Name, *d.PinnedAt), true
	}
	return *d.Name, true
}

type Digester interface{ Digest(string) (string, error) }

type CraneDigester struct{}

func (c CraneDigester) Digest(name string) (string, error) {
	//nolint:wrapcheck // error value not used
	return crane.Digest(name)
}

// CreateDockerfilePinningRemediation create remediaiton for pinning Dockerfile images.
func CreateDockerfilePinningRemediation(dep *checker.Dependency, digester Digester) *checker.Remediation {
	name, ok := dockerImageName(dep)
	if !ok {
		return nil
	}

	hash, err := digester.Digest(name)
	if err != nil {
		return nil
	}

	text := fmt.Sprintf(dockerfilePinText, name, hash)
	markdown := text

	return &checker.Remediation{
		HelpText:     text,
		HelpMarkdown: markdown,
	}
}
