// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

type Repo struct {
	scheme        string
	host          string
	organization  string
	project       string
	projectID     string
	name          string
	id            string
	defaultBranch string
	commitSHA     string
	metadata      []string
}

const (
	azureDevOpsHost        = "dev.azure.com"
	visualStudioHostSuffix = ".visualstudio.com"
	defaultCollection      = "DefaultCollection"
	gitPathSegment         = "_git"
)

var errInvalidAzureDevOpsPathSegment = errors.New("invalid Azure DevOps path segment")

func hasAzureDevOpsHost(input string) bool {
	u, err := url.Parse(withDefaultScheme(input))
	if err != nil {
		return false
	}

	host := strings.ToLower(u.Hostname())
	return host == azureDevOpsHost || strings.HasSuffix(host, visualStudioHostSuffix)
}

// HasAzureDevOpsHost reports whether input uses an Azure DevOps cloud hostname.
func HasAzureDevOpsHost(input string) bool {
	return hasAzureDevOpsHost(input)
}

func pathSegments(u *url.URL) ([]string, error) {
	escapedSegments := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	segments := make([]string, len(escapedSegments))
	for i, escapedSegment := range escapedSegments {
		segment, err := url.PathUnescape(escapedSegment)
		if err != nil || strings.Contains(segment, "/") {
			return nil, errInvalidAzureDevOpsPathSegment
		}
		segments[i] = segment
	}
	return segments, nil
}

// Parses input string into repoURL struct
/*
 Accepted input string formats, with or without an https scheme, are as follows:
	- "dev.azure.com/<organization:string>/<project:string>/_git/<repository:string>"
	- "<organization:string>.visualstudio.com/<project:string>/_git/<repository:string>"
	- "<organization:string>.visualstudio.com/DefaultCollection/<project:string>/_git/<repository:string>"
*/
func (r *Repo) parse(input string) error {
	u, err := url.Parse(withDefaultScheme(input))
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", err))
	}

	if !strings.EqualFold(u.Scheme, "https") || u.Port() != "" {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Azure DevOps repo format is invalid: %s", input))
	}

	host := strings.ToLower(u.Hostname())
	segments, err := pathSegments(u)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Azure DevOps repo format is invalid: %s", input))
	}

	var organization, project, gitSegment, name string
	switch {
	case host == azureDevOpsHost:
		if len(segments) != 4 {
			return sce.WithMessage(
				sce.ErrScorecardInternal,
				fmt.Sprintf("Azure DevOps repo format is invalid: %s", input),
			)
		}
		organization, project, gitSegment, name = segments[0], segments[1], segments[2], segments[3]
	case strings.HasSuffix(host, visualStudioHostSuffix):
		organization = strings.TrimSuffix(host, visualStudioHostSuffix)
		if organization == "" || strings.Contains(organization, ".") {
			return sce.WithMessage(
				sce.ErrScorecardInternal,
				fmt.Sprintf("Azure DevOps repo format is invalid: %s", input),
			)
		}

		switch {
		case len(segments) == 3:
			project, gitSegment, name = segments[0], segments[1], segments[2]
		case len(segments) == 4 && strings.EqualFold(segments[0], defaultCollection):
			project, gitSegment, name = segments[1], segments[2], segments[3]
		default:
			return sce.WithMessage(
				sce.ErrScorecardInternal,
				fmt.Sprintf("Azure DevOps repo format is invalid: %s", input),
			)
		}
	default:
		return sce.WithMessage(
			sce.ErrScorecardInternal,
			fmt.Sprintf("Azure DevOps repo format is invalid: %s", input),
		)
	}

	if organization == "" || project == "" || !strings.EqualFold(gitSegment, gitPathSegment) || name == "" {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Azure DevOps repo format is invalid: %s", input))
	}

	r.scheme = "https"
	r.host = azureDevOpsHost
	r.organization = organization
	r.project = project
	r.name = name
	return nil
}

// Allow skipping scheme for ease-of-use, default to https.
func withDefaultScheme(uri string) string {
	if strings.Contains(uri, "://") {
		return uri
	}
	return "https://" + uri
}

// URI implements Repo.URI().
func (r *Repo) URI() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", r.host, r.organization, r.project, "_git", r.name)
}

func (r *Repo) Host() string {
	return r.host
}

// String implements Repo.String.
func (r *Repo) String() string {
	return fmt.Sprintf("%s-%s_%s_%s", r.host, r.organization, r.project, r.name)
}

// IsValid checks if the repoURL is valid.
func (r *Repo) IsValid() error {
	if strings.TrimSpace(r.organization) == "" ||
		strings.TrimSpace(r.project) == "" ||
		strings.TrimSpace(r.name) == "" {
		return sce.WithMessage(sce.ErrInvalidURL, "expected full project url: "+r.URI())
	}

	return nil
}

func (r *Repo) AppendMetadata(metadata ...string) {
	r.metadata = append(r.metadata, metadata...)
}

// Metadata implements Repo.Metadata.
func (r *Repo) Metadata() []string {
	return r.metadata
}

// Type implements Repo.Type.
func (r *Repo) Type() clients.RepoType {
	return clients.RepoTypeAzureDevOps
}

// Path() implements RepoClient.Path.
func (r *Repo) Path() string {
	return fmt.Sprintf("%s/%s/%s/%s", r.organization, r.project, "_git", r.name)
}

// MakeAzureDevOpsRepo takes input of forms in parse and returns and implementation
// of clients.Repo interface.
func MakeAzureDevOpsRepo(input string) (clients.Repo, error) {
	var repo Repo
	if err := repo.parse(input); err != nil {
		return nil, fmt.Errorf("error during parse: %w", err)
	}
	if err := repo.IsValid(); err != nil {
		return nil, fmt.Errorf("error in IsValid: %w", err)
	}

	return &repo, nil
}
