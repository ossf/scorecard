// Copyright 2026 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"net/url"
	"os"
	"strings"
)

// tokenForHost returns GITLAB_AUTH_TOKEN only for the GitLab instance the
// token belongs to: GL_HOST when it is set, otherwise gitlab.com. Every other
// host gets an empty token, so requests to it are unauthenticated.
func tokenForHost(host string) string {
	if strings.EqualFold(host, tokenHost()) {
		return os.Getenv("GITLAB_AUTH_TOKEN")
	}
	return ""
}

// tokenHost returns the host GITLAB_AUTH_TOKEN may be sent to, in the same
// form Repo.Host reports: GL_HOST's host plus any path prefix.
func tokenHost() string {
	h := os.Getenv("GL_HOST")
	if h == "" {
		return "gitlab.com"
	}
	u, err := url.Parse(withDefaultScheme(h))
	if err != nil {
		return ""
	}
	return u.Host + strings.TrimRight(u.Path, "/")
}
