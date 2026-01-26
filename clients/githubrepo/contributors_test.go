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
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v53/github"

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
			sb.WriteString(fmt.Sprintf(" @user%d", i))
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
