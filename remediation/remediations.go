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

	"github.com/google/go-containerregistry/pkg/crane"

	"github.com/ossf/scorecard/v4/clients"
)

var (
	workflowText = "update your workflow using https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s"
	//nolint
	workflowMarkdown  = "update your workflow using [https://app.stepsecurity.io](https://app.stepsecurity.io/secureworkflow/%s/%s/%s?enable=%s)"
	dockerfilePinText = "pin your Docker image by updating %[1]s to %[1]s@%s"

	errInvalidArg = errors.New("invalid argument")
)

type MetadataKey string

const (
	BranchName MetadataKey = "branch"
	RepoName   MetadataKey = "repo"
)

type Metadata map[MetadataKey]string

func extractRepoName(rc clients.RepoClient) (string, error) {
	uri := rc.URI()
	parts := strings.Split(uri, "/")
	if len(parts) != 3 {
		return "", errInvalidArg
	}
	return fmt.Sprintf("%s/%s", parts[1], parts[2]), nil
}

func NewMetadata(rc clients.RepoClient) Metadata {
	md := Metadata{}
	if rc == nil {
		return md
	}

	if branch, err := rc.GetDefaultBranchName(); err == nil {
		md[BranchName] = branch
	}

	if repo, err := extractRepoName(rc); err == nil {
		md[RepoName] = repo
	}

	return md
}

// Remediation represents a remediation.
type Remediation struct {
	// Code snippet for humans.
	Snippet string
	// Diff for machines.
	Diff string
	// Help text for humans.
	HelpText string
	// Help text in markdown format for humans.
	HelpMarkdown string
}

// CreateWorkflowPermissionRemediation create remediation for workflow permissions.
func CreateWorkflowPermissionRemediation(md Metadata, filepath string) *Remediation {
	return createWorkflowRemediation(md, filepath, "permissions")
}

// CreateWorkflowPinningRemediation create remediaiton for pinninn GH Actions.
func CreateWorkflowPinningRemediation(md Metadata, filepath string) *Remediation {
	return createWorkflowRemediation(md, filepath, "pin")
}

func createWorkflowRemediation(md Metadata, path, t string) *Remediation {
	p := strings.TrimPrefix(path, ".github/workflows/")

	branch, bOk := md[BranchName]
	repo, rOk := md[RepoName]
	if !bOk || !rOk {
		return nil
	}

	text := fmt.Sprintf(workflowText, repo, p, branch, t)
	markdown := fmt.Sprintf(workflowMarkdown, repo, p, branch, t)

	return &Remediation{
		HelpText:     text,
		HelpMarkdown: markdown,
	}
}

// CreateDockerfilePinningRemediation create remediaiton for pinning Dockerfile images.
func CreateDockerfilePinningRemediation(md Metadata, name *string) *Remediation {
	if name == nil {
		return nil
	}
	hash, err := crane.Digest(*name)
	if err != nil {
		return nil
	}

	text := fmt.Sprintf(dockerfilePinText, *name, hash)
	markdown := text

	return &Remediation{
		HelpText:     text,
		HelpMarkdown: markdown,
	}
}
