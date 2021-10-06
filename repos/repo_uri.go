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
	// ErrorUnsupportedhost indicates the repo's host is unsupported.
	ErrorUnsupportedhost = errors.New("unsupported host")
	// ErrorInvalidGithubURL indicates the repo's GitHub URL is not in the proper format.
	ErrorInvalidGithubURL = errors.New("invalid GitHub repo URL")
	// ErrorInvalidURL indicates the repo's full GitHub URL was not passed.
	ErrorInvalidURL = errors.New("invalid repo flag")
)

// RepoURI represents the URI for a repo.
type RepoURI struct {
	repoType RepoType
	url      repoURL
	localDir repoLocalDir
	metadata []string
}

type repoLocalDir struct {
	path string
}

type repoURL struct {
	host, owner, repo string
}

// RepoType is the type of a file.
type RepoType int

const (
	// RepoTypeURL is for URLs.
	RepoTypeURL RepoType = iota
	// RepoTypeLocalDir is for source code in directories.
	RepoTypeLocalDir
)

// NewFromURL creates a RepoURI from URL.
func NewFromURL(url string) (*RepoURI, error) {
	r := &RepoURI{
		repoType: RepoTypeURL,
	}

	if err := r.Set(url); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return r, nil
}

// NewFromLocalDirectory creates a RepoURI as a local directory.
func NewFromLocalDirectory(path string) *RepoURI {
	return &RepoURI{
		localDir: repoLocalDir{
			path: path,
		},
		repoType: RepoTypeLocalDir,
	}
}

func (r *RepoURI) SetMetadata(m []string) error {
	r.metadata = m
	return nil
}

func (r *RepoURI) SetURL(url string) error {
	if err := r.Set(url); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// Type method is needed so that this struct can be used as cmd flag.
func (r *RepoURI) Type() string {
	return "repo"
}

// GetType gives the type of URI.
func (r *RepoURI) GetType() RepoType {
	return r.repoType
}

// GetPath retusn the path for a local directory.
func (r *RepoURI) GetPath() string {
	return r.localDir.path
}

// GetURL returns a valid url for Repo struct.
func (r *RepoURI) GetURL() string {
	return fmt.Sprintf("%s/%s/%s", r.url.host, r.url.owner, r.url.repo)
}

// GetMetadata returns a valid url for Repo struct.
func (r *RepoURI) GetMetadata() []string {
	return r.metadata
}

// String returns a string representation of Repo struct.
func (r *RepoURI) String() string {
	return fmt.Sprintf("%s-%s-%s", r.url.host, r.url.owner, r.url.repo)
}

// Set parses a URI string into Repo struct.
func (r *RepoURI) Set(s string) error {
	const two = 2
	const three = 3

	// TODO: can we use this for pypi://, rubygems://, etc?
	const httpsPrefix = "https://"
	const filePrefix = "file://"

	// Validate the URI and scheme.
	if !strings.HasPrefix(s, filePrefix) &&
		!strings.HasPrefix(s, httpsPrefix) {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid URI: %v", s))
	}

	// Parse the URI.
	// c = strings.Split(s, "/")

	// switch l := len(c); {
	// // This will takes care for repo/url.owner format.
	// // By default it will use github.com
	// case l == two:
	// 	t = "github.com/" + c[0] + "/" + c[1]
	// case l >= three:
	// 	t = s
	// }

	// // Allow skipping scheme for ease-of-use, default to https.
	// if !strings.Contains(t, "://") {
	// 	t = "https://" + t
	// }
	// fmt.Println(s)

	u, e := url.Parse(s)
	if e != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", e))
	}

	switch {
	case strings.HasPrefix(s, httpsPrefix):
		const splitLen = 2
		split := strings.SplitN(strings.Trim(u.Path, "/"), "/", splitLen)
		if len(split) != splitLen {
			return sce.WithMessage(ErrorInvalidURL, fmt.Sprintf("%v. Expected full repository url", s))
		}
		r.url.host, r.url.owner, r.url.repo = u.Host, split[0], split[1]
	case strings.HasPrefix(s, filePrefix):
		r.localDir.path = s[len(filePrefix):]
		r.repoType = RepoTypeLocalDir
	default:
		break
	}

	return nil
}

// IsValidGitHubURL checks whether Repo represents a valid GitHub repo and returns errors otherwise.
func (r *RepoURI) IsValidGitHubURL() error {
	switch r.url.host {
	case "github.com":
	default:
		return sce.WithMessage(ErrorUnsupportedhost, r.url.host)
	}

	if strings.TrimSpace(r.url.owner) == "" || strings.TrimSpace(r.url.repo) == "" {
		return sce.WithMessage(ErrorInvalidGithubURL,
			fmt.Sprintf("%v. Expected the full reposiroty url", r.GetURL()))
	}
	return nil
}
