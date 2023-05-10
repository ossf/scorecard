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

// NOTE: In Gitlab repositories are called projects, however to ensure compatibility,
// this package will regard to Gitlab projects as repositories.
package gitlabrepo

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type repoURL struct {
	scheme        string
	host          string
	owner         string
	project       string
	defaultBranch string
	commitSHA     string
	metadata      []string
}

var errInvalidGitlabRepoURL = errors.New("repo is not a gitlab repo")

// Parses input string into repoURL struct
/*
*  Accepted input string formats are as follows:
	*  "gitlab.<companyDomain:string>.com/<owner:string>/<projectID:string>"
	* "https://gitlab.<companyDomain:string>.com/<owner:string>/<projectID:string>"

The following input format is not supported:
	* https://gitlab.<companyDomain:string>.com/projects/<projectID:int>
*/
func (r *repoURL) parse(input string) error {
	var t string
	c := strings.Split(input, "/")
	switch l := len(c); {
	// owner/repo format is not supported for gitlab, it's github-only
	case l == 2:
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("gitlab repo must specify host: %s", input))
	case l >= 3:
		t = input
	}

	// Allow skipping scheme for ease-of-use, default to https.
	if !strings.Contains(t, "://") {
		t = "https://" + t
	}

	u, e := url.Parse(t)
	if e != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", e))
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(u.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		return sce.WithMessage(sce.ErrorInvalidURL, fmt.Sprintf("%v. Expected full repository url", input))
	}

	r.scheme, r.host, r.owner, r.project = u.Scheme, u.Host, split[0], split[1]
	return nil
}

// URI implements Repo.URI().
func (r *repoURL) URI() string {
	return fmt.Sprintf("%s/%s/%s", r.host, r.owner, r.project)
}

func (r *repoURL) Host() string {
	return fmt.Sprintf("%s://%s", r.scheme, r.host)
}

// String implements Repo.String.
func (r *repoURL) String() string {
	return fmt.Sprintf("%s-%s_%s", r.host, r.owner, r.project)
}

// IsValid implements Repo.IsValid.
func (r *repoURL) IsValid() error {
	if strings.Contains(r.host, "gitlab.") {
		return nil
	}

	if strings.EqualFold(r.host, "github.com") {
		return fmt.Errorf("%w: %s", errInvalidGitlabRepoURL, r.host)
	}

	client, err := gitlab.NewClient("", gitlab.WithBaseURL(fmt.Sprintf("%s://%s", r.scheme, r.host)))
	if err != nil {
		return sce.WithMessage(err,
			fmt.Sprintf("couldn't create gitlab client for %s", r.host),
		)
	}

	_, resp, err := client.Projects.ListProjects(&gitlab.ListProjectsOptions{})
	if resp == nil || resp.StatusCode != 200 {
		return sce.WithMessage(sce.ErrRepoUnreachable,
			fmt.Sprintf("couldn't reach gitlab instance at %s", r.host),
		)
	}
	if err != nil {
		return sce.WithMessage(err,
			fmt.Sprintf("error when connecting to gitlab instance at %s", r.host),
		)
	}

	if strings.TrimSpace(r.owner) == "" || strings.TrimSpace(r.project) == "" {
		return sce.WithMessage(sce.ErrorInvalidURL,
			fmt.Sprintf("%v. Expected the full project url", r.URI()))
	}
	return nil
}

func (r *repoURL) AppendMetadata(metadata ...string) {
	r.metadata = append(r.metadata, metadata...)
}

// Metadata implements Repo.Metadata.
func (r *repoURL) Metadata() []string {
	return r.metadata
}

// MakeGitlabRepo takes input of forms in parse and returns and implementation
// of clients.Repo interface.
func MakeGitlabRepo(input string) (clients.Repo, error) {
	var repo repoURL
	if err := repo.parse(input); err != nil {
		return nil, fmt.Errorf("error during parse: %w", err)
	}
	if err := repo.IsValid(); err != nil {
		return nil, fmt.Errorf("error in IsValid: %w", err)
	}
	return &repo, nil
}
