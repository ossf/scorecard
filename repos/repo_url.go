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

// Package repos defines a generic repository.
package repos

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	sce "github.com/ossf/scorecard/v3/errors"
)

var (
	// ErrorUnsupportedHost indicates the repo's host is unsupported.
	ErrorUnsupportedHost = errors.New("unsupported host")
	// ErrorInvalidGithubURL indicates the repo's GitHub URL is not in the proper format.
	ErrorInvalidGithubURL = errors.New("invalid GitHub repo URL")
	// ErrorInvalidURL indicates the repo's full GitHub URL was not passed.
	ErrorInvalidURL = errors.New("invalid repo flag")
)

// nolint:revive
// TODO: Remove RepoURL and replace all instances with clients.Repo interface.
type RepoURL struct {
	Host, Owner, Repo string
	Metadata          []string
}

// Type method is needed so that this struct can be used as cmd flag.
func (r *RepoURL) Type() string {
	return "repo"
}

// URL returns a valid url for RepoURL struct.
func (r *RepoURL) URL() string {
	return fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
}

// String returns a string representation of RepoURL struct.
func (r *RepoURL) String() string {
	return fmt.Sprintf("%s-%s-%s", r.Host, r.Owner, r.Repo)
}

// Set parses a URL string into RepoURL struct.
func (r *RepoURL) Set(s string) error {
	var t string

	const two = 2
	const three = 3

	c := strings.Split(s, "/")

	switch l := len(c); {
	// This will takes care for repo/owner format.
	// By default it will use github.com
	case l == two:
		t = "github.com/" + c[0] + "/" + c[1]
	case l >= three:
		t = s
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
		return sce.WithMessage(ErrorInvalidURL, fmt.Sprintf("%v. Exepted full repository url", s))
	}

	r.Host, r.Owner, r.Repo = u.Host, split[0], split[1]
	return nil
}

// ValidGitHubURL checks whether RepoURL represents a valid GitHub repo and returns errors otherwise.
func (r *RepoURL) ValidGitHubURL() error {
	switch r.Host {
	case "github.com":
	default:
		return sce.WithMessage(ErrorUnsupportedHost, r.Host)
	}

	if strings.TrimSpace(r.Owner) == "" || strings.TrimSpace(r.Repo) == "" {
		return sce.WithMessage(ErrorInvalidGithubURL,
			fmt.Sprintf("%v. Expected the full reposiroty url", r.URL()))
	}
	return nil
}
