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

// NOTE: In GitLab repositories are called projects, however to ensure compatibility,
// this package will regard to GitLab projects as repositories.
package gitlabrepo

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

type repoURL struct {
	scheme        string
	host          string
	owner         string
	project       string
	projectID     string
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

	u, err := url.Parse(withDefaultScheme(t))
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", err))
	}

	// fixup the URL, for situations where GL_HOST contains part of the path
	// https://github.com/ossf/scorecard/issues/3696
	if h := os.Getenv("GL_HOST"); h != "" {
		hostURL, err := url.Parse(withDefaultScheme(h))
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse GL_HOST: %v", err))
		}

		// only modify behavior of repos which fall under GL_HOST
		if hostURL.Host == u.Host {
			// without the scheme and without trailing slashes
			u.Host = hostURL.Host + strings.TrimRight(hostURL.Path, "/")
			// remove any part of the path which belongs to the host
			u.Path = strings.TrimPrefix(u.Path, hostURL.Path)
		}
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(u.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		return sce.WithMessage(sce.ErrInvalidURL, fmt.Sprintf("%v. Expected full repository url", input))
	}

	r.scheme, r.host, r.owner, r.project = u.Scheme, u.Host, split[0], split[1]
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
func (r *repoURL) URI() string {
	return fmt.Sprintf("%s/%s/%s", r.host, r.owner, r.project)
}

func (r *repoURL) Host() string {
	return r.host
}

// String implements Repo.String.
func (r *repoURL) String() string {
	return fmt.Sprintf("%s-%s_%s", r.host, r.owner, r.project)
}

// IsValid implements Repo.IsValid.
func (r *repoURL) IsValid() error {
	if strings.TrimSpace(r.owner) == "" || strings.TrimSpace(r.project) == "" {
		return sce.WithMessage(sce.ErrInvalidURL, "expected full project url: "+r.URI())
	}

	if strings.Contains(r.host, "gitlab.") {
		return nil
	}

	if strings.EqualFold(r.host, "github.com") {
		return fmt.Errorf("%w: %s", errInvalidGitlabRepoURL, r.host)
	}

	// intentionally pass empty token
	// "When accessed without authentication, only public projects with simple fields are returned."
	// https://docs.gitlab.com/ee/api/projects.html#list-all-projects
	baseURL := fmt.Sprintf("%s://%s", r.scheme, r.host)
	client, err := gitlab.NewClient("", gitlab.WithBaseURL(baseURL))
	if err != nil {
		return sce.WithMessage(err,
			fmt.Sprintf("couldn't create gitlab client for %s", r.host),
		)
	}

	_, resp, err := client.Projects.ListProjects(&gitlab.ListProjectsOptions{})
	if resp == nil || resp.StatusCode != http.StatusOK {
		return sce.WithMessage(sce.ErrRepoUnreachable,
			fmt.Sprintf("couldn't reach gitlab instance at %s: %v", r.host, err),
		)
	}
	if err != nil {
		return sce.WithMessage(err,
			fmt.Sprintf("error when connecting to gitlab instance at %s", r.host),
		)
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

// Path() implements RepoClient.Path.
func (r *repoURL) Path() string {
	return fmt.Sprintf("%s/%s", r.owner, r.project)
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
