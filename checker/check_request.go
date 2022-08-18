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

package checker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/clients"
)

var errInvalidArg = errors.New("invalid argument")

// CheckRequest struct encapsulates all data to be passed into a CheckFn.
type CheckRequest struct {
	Ctx                   context.Context
	RepoClient            clients.RepoClient
	CIIClient             clients.CIIBestPracticesClient
	OssFuzzRepo           clients.RepoClient
	Dlogger               DetailLogger
	Repo                  clients.Repo
	VulnerabilitiesClient clients.VulnerabilitiesClient
	// UPGRADEv6: return raw results instead of scores.
	RawResults    *RawResults
	RequiredTypes []RequestType
}

// RequestType identifies special requirements/attributes that need to be supported by checks.
type RequestType int

const (
	// FileBased request types require checks to run solely on file-content.
	FileBased RequestType = iota
	// CommitBased request types require checks to run on non-HEAD commit content.
	CommitBased
)

// ListUnsupported returns []RequestType not in `supported` and are `required`.
func ListUnsupported(required, supported []RequestType) []RequestType {
	var ret []RequestType
	for _, t := range required {
		if !contains(supported, t) {
			ret = append(ret, t)
		}
	}
	return ret
}

func contains(in []RequestType, exists RequestType) bool {
	for _, r := range in {
		if r == exists {
			return true
		}
	}
	return false
}

// remediationMetadata returns remediation relevant metadata from a CheckRequest.
func (c *CheckRequest) SetRemediationMetadata() error {
	if c == nil || c.RepoClient == nil || c.RawResults == nil {
		return nil
	}

	// Get the branch for remediation.
	branch, err := c.RepoClient.GetDefaultBranchName()
	if err != nil {
		return fmt.Errorf("GetDefaultBranchName: %w", err)
	}

	uri := c.RepoClient.URI()
	parts := strings.Split(uri, "/")
	if len(parts) != 3 {
		return fmt.Errorf("%w: empty: %s", errInvalidArg, uri)
	}
	repo := fmt.Sprintf("%s/%s", parts[1], parts[2])
	c.RawResults.RemediationMetadata = RemediationMetadata{Branch: branch, Repo: repo}
	return nil
}
