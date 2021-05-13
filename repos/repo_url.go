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

package repos

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrorUnsupportedHost  = errors.New("unsupported host")
	ErrorInvalidGithubURL = errors.New("invalid GitHub repo URL")
	ErrorInvalidURL       = errors.New("invalid repo flag")
)

type RepoURL struct {
	Host, Owner, Repo string
}

// Type method is needed so that this struct can be used as cmd flag.
func (r *RepoURL) Type() string {
	return "repo"
}

func (r *RepoURL) URL() string {
	return fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) String() string {
	return fmt.Sprintf("%s-%s-%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) Set(s string) error {
	// Allow skipping scheme for ease-of-use, default to https.
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	u, e := url.Parse(s)
	if e != nil {
		return fmt.Errorf("error parsing repo URL: %w", e)
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(u.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		return fmt.Errorf("%w: [%s], pass the full repository URL", ErrorInvalidURL, s)
	}

	r.Host, r.Owner, r.Repo = u.Host, split[0], split[1]
	return nil
}

func (r *RepoURL) ValidGitHubURL() error {
	switch r.Host {
	case "github.com":
		break
	default:
		return fmt.Errorf("%w: %s", ErrorUnsupportedHost, r.Host)
	}

	if strings.TrimSpace(r.Owner) == "" || strings.TrimSpace(r.Repo) == "" {
		return fmt.Errorf("%w: [%s], pass the full repository URL", ErrorInvalidGithubURL, r.URL())
	}
	return nil
}
