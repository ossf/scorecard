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

// Parses input string into repoURL struct
/*
 Accepted input string formats are as follows:
	- "dev.azure.com/<organization:string>/<project:string>/_git/<repository:string>"
	- "https://dev.azure.com/<organization:string>/<project:string>/_git/<repository:string>"
*/
func (r *Repo) parse(input string) error {
	u, err := url.Parse(withDefaultScheme(input))
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", err))
	}

	const splitLen = 4
	split := strings.SplitN(strings.Trim(u.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Azure DevOps repo format is invalid: %s", input))
	}

	r.scheme, r.host, r.organization, r.project, r.name = u.Scheme, u.Host, split[0], split[1], split[3]
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
