// Copyright 2025 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package githubrepo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients"
)

type codeownerRoundTripper struct{}

const tooManyCodeowners = 105

func (c codeownerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Path {
	case "/orgs/ossf-tests/teams/foo/members":
		type member struct {
			Login string `json:"login"`
		}
		members := make([]member, 0, tooManyCodeowners)
		for i := range tooManyCodeowners {
			members = append(members, member{Login: fmt.Sprintf("user%d", i)})
		}
		jsonResp, err := json.Marshal(members)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(jsonResp)),
		}, nil
	default:
		return nil, errors.New("unsupported URL")
	}
}

func Test_mapCodeOwners(t *testing.T) {
	t.Parallel()
	httpClient := &http.Client{
		Transport: codeownerRoundTripper{},
	}
	client := github.NewClient(httpClient)
	handler := &contributorsHandler{
		ghClient: client,
		ctx:      t.Context(),
	}
	t.Run("one team with more too many codeowners", func(t *testing.T) {
		t.Parallel()
		codeowners := io.NopCloser(strings.NewReader("* @ossf-tests/foo\n"))
		contributors := map[string]clients.User{}
		mapCodeOwners(handler, codeowners, contributors)
		got := len(contributors)
		if got != codeownersLimit {
			t.Errorf("wanted less than %d CODEOWNERs, got %d", codeownersLimit, got)
		}
	})

	t.Run("too many individual codeowners", func(t *testing.T) {
		t.Parallel()
		var sb strings.Builder
		sb.WriteRune('*')
		for i := range tooManyCodeowners {
			fmt.Fprintf(&sb, " @user%d", i)
		}
		sb.WriteString("\n")
		codeowners := io.NopCloser(strings.NewReader(sb.String()))
		contributors := map[string]clients.User{}
		mapCodeOwners(handler, codeowners, contributors)
		got := len(contributors)
		if got != codeownersLimit {
			t.Errorf("wanted less than %d CODEOWNERs, got %d", codeownersLimit, got)
		}
	})
}

// teamsRoundTripper serves team member lists and records each team
// whose members are requested.
type teamsRoundTripper struct {
	members  map[string][]string
	requests []string
	mu       sync.Mutex
}

func (rt *teamsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.mu.Lock()
	rt.requests = append(rt.requests, req.URL.Path)
	rt.mu.Unlock()
	members, ok := rt.members[req.URL.Path]
	if !ok {
		return &http.Response{StatusCode: http.StatusForbidden, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	var body strings.Builder
	body.WriteString("[")
	for i, login := range members {
		if i > 0 {
			body.WriteString(",")
		}
		fmt.Fprintf(&body, `{"login": %q}`, login)
	}
	body.WriteString("]")
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body.String()))}, nil
}

func Test_mapCodeOwners_contributors(t *testing.T) {
	t.Parallel()
	const (
		smallTeam = "/orgs/ossf-tests/teams/small/members"
		largeTeam = "/orgs/ossf-tests/teams/large/members"
		denied    = "/orgs/ossf-tests/teams/denied/members"
	)
	var individuals strings.Builder
	individuals.WriteString("/a/")
	for i := range codeownersLimit - 1 {
		fmt.Fprintf(&individuals, " @user%d", i)
	}
	large := make([]string, 0, 3)
	for i := range 3 {
		large = append(large, fmt.Sprintf("member%d", i))
	}
	tests := []struct {
		existing map[string]clients.User
		want     map[string]clients.User
		name     string
		file     string
		requests []string
		nilFile  bool
	}{
		{
			name:     "no CODEOWNERS file",
			nilFile:  true,
			existing: map[string]clients.User{"alice": {Login: "alice", NumContributions: 5}},
			want:     map[string]clients.User{"alice": {Login: "alice", NumContributions: 5}},
		},
		{
			name:     "unparsable CODEOWNERS file",
			file:     "a[ @alice\n",
			existing: map[string]clients.User{"alice": {Login: "alice", NumContributions: 5}},
			want:     map[string]clients.User{"alice": {Login: "alice", NumContributions: 5}},
		},
		{
			name: "users, teams, and emails",
			file: "* @alice @ossf-tests/small\n" +
				"/docs/ @bob user@example.com @ossf-tests/denied\n" +
				"/x/ @alice @carol\n",
			existing: map[string]clients.User{
				"alice": {Login: "alice", NumContributions: 5},
				"dave":  {Login: "dave", NumContributions: 1},
			},
			// The small team has a member with an empty login, which is skipped.
			want: map[string]clients.User{
				"alice": {Login: "alice", NumContributions: 5, IsCodeOwner: true},
				"dave":  {Login: "dave", NumContributions: 1},
				"bob":   {Login: "bob", IsCodeOwner: true},
				"carol": {Login: "carol", IsCodeOwner: true},
				"erin":  {Login: "erin", IsCodeOwner: true},
			},
			requests: []string{smallTeam, denied},
		},
		{
			name: "limit reached inside a team",
			file: individuals.String() + "\n" +
				"/b/ @ossf-tests/large @ossf-tests/small\n" +
				"/c/ @zed\n",
			want: func() map[string]clients.User {
				want := map[string]clients.User{}
				for i := range codeownersLimit - 1 {
					login := fmt.Sprintf("user%d", i)
					want[login] = clients.User{Login: login, IsCodeOwner: true}
				}
				want["member0"] = clients.User{Login: "member0", IsCodeOwner: true}
				return want
			}(),
			requests: []string{largeTeam},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rt := &teamsRoundTripper{members: map[string][]string{
				smallTeam: {"erin", "", "bob"},
				largeTeam: large,
			}}
			handler := &contributorsHandler{
				ghClient: github.NewClient(&http.Client{Transport: rt}),
				ctx:      t.Context(),
			}
			var file io.ReadCloser
			if !tt.nilFile {
				file = io.NopCloser(strings.NewReader(tt.file))
			}
			contributors := maps.Clone(tt.existing)
			if contributors == nil {
				contributors = map[string]clients.User{}
			}

			mapCodeOwners(handler, file, contributors)

			if diff := cmp.Diff(tt.want, contributors); diff != "" {
				t.Errorf("contributors mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.requests, rt.requests); diff != "" {
				t.Errorf("team requests mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
