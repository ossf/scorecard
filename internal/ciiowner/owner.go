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

package ciiowner

import (
	"net/url"
	"strings"
)

// SameProject reports whether candidateRepoURL may inherit the OpenSSF badge
// registered for badgeRepoURL. Both must share the same host and owner namespace.
//
// The namespace is everything in the URL path except the final repo-name segment,
// so github.com/org/repo-a and github.com/org/repo-b share the namespace "org",
// while gitlab.com/g/sub1/repo and gitlab.com/g/sub2/repo do not.
func SameProject(candidateRepoURL, badgeRepoURL string) bool {
	c := normalize(candidateRepoURL)
	b := normalize(badgeRepoURL)
	if c == nil || b == nil {
		return false
	}
	if c.Host != b.Host {
		return false
	}
	ns := namespace(c)
	return ns != "" && ns == namespace(b)
}

// normalize parses and canonicalises a repository URL.
// Returns nil for malformed or userinfo-bearing URLs.
func normalize(raw string) *url.URL {
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil {
		return nil
	}
	hostname := strings.ToLower(u.Hostname())
	if port := u.Port(); port != "" && port != "80" && port != "443" {
		hostname += ":" + port
	}
	path := strings.ToLower(u.Path)
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimSuffix(path, ".git")
	return &url.URL{Host: hostname, Path: path}
}

// namespace returns the owner/group prefix of a repository URL — the full path
// minus the final segment (the repo name). Returns "" for single-segment or
// empty paths so that bare org URLs are rejected.
func namespace(u *url.URL) string {
	path := strings.TrimPrefix(u.Path, "/")
	idx := strings.LastIndex(path, "/")
	if idx < 1 {
		return ""
	}
	return path[:idx]
}
