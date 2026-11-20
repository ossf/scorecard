// Copyright 2023 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/shurcooL/githubv4"

	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/log"
)

func Test_init_clearsErr(t *testing.T) {
	t.Parallel()
	handler := &checkrunsHandler{errSetup: sce.ErrScorecardInternal}
	handler.init(t.Context(), nil, 0)
	if handler.errSetup != nil {
		t.Errorf("expected nil error, got %v", handler.errSetup)
	}
}

// graphqlMessageRoundTripper returns a GraphQL-level error response body with
// the given message, simulating errors such as GitHub org IP allow-list rejections.
type graphqlMessageRoundTripper struct {
	message string
}

func (g graphqlMessageRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	body := fmt.Sprintf(`{"errors":[{"message":%q}]}`, g.message)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}, nil
}

func newTestCheckrunsHandler(t *testing.T, rt http.RoundTripper) *checkrunsHandler {
	t.Helper()
	return &checkrunsHandler{
		graphClient:    githubv4.NewClient(&http.Client{Transport: rt}),
		repourl:        &Repo{owner: "ossf", repo: "scorecard"},
		logger:         log.NewLogger(log.DefaultLevel),
		checkData:      new(checkRunsGraphqlData),
		setupOnce:      new(sync.Once),
		checkRunsByRef: checkRunsByRef{},
		commitDepth:    1,
		ctx:            t.Context(),
	}
}

// Test_setup_ipAllowListError verifies that when the GraphQL query fails with an
// org IP allow-list error, setup() does NOT propagate the error (errSetup stays nil),
// so CI-Tests and SAST checks degrade gracefully instead of aborting entirely.
// Regression test for https://github.com/ossf/scorecard/issues/5382
func Test_setup_ipAllowListError(t *testing.T) {
	t.Parallel()

	rt := graphqlMessageRoundTripper{
		message: "Although you appear to have the correct authorization credentials, " +
			"the `mistralai` organization has an IP allow list enabled, " +
			"and your IP address is not permitted to access this resource.",
	}
	handler := newTestCheckrunsHandler(t, rt)

	err := handler.setup()
	if err != nil {
		t.Errorf("expected nil error for IP allow-list rejection, got: %v", err)
	}
}

// Test_setup_unrelatedErrorPropagates verifies that genuine GraphQL errors
// (not IP allow-list or integration-access errors) are still propagated as before.
func Test_setup_unrelatedErrorPropagates(t *testing.T) {
	t.Parallel()

	rt := graphqlMessageRoundTripper{
		message: "some unexpected internal server error",
	}
	handler := newTestCheckrunsHandler(t, rt)

	err := handler.setup()
	if err == nil {
		t.Error("expected error to be propagated for unrelated GraphQL errors, got nil")
	}
}
