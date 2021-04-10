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
package pkg

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// TODO: This is an exact replica of repos.RepoURL. Combine them somehow.
// RepoURL parses and stores URL into fields.
type RepoURL struct {
	Host  string // Host where the repo is stored. Example GitHub.com
	Owner string // Owner of the repo. Example ossf.
	Repo  string // The actual repo. Example scorecard.
}

func (r *RepoURL) String() string {
	return fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) NonURLString() string {
	return fmt.Sprintf("%s-%s-%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) Set(s string) error {
	// Allow skipping scheme for ease-of-use, default to https.
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	parsedURL, err := url.Parse(s)
	if err != nil {
		return errors.Wrap(err, "unable to parse the URL")
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(parsedURL.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		return errors.Errorf("invalid repo flag: [%s], pass the full repository URL", s)
	}

	r.Host, r.Owner, r.Repo = parsedURL.Host, split[0], split[1]
	return nil
}
