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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v82/github"

	sce "github.com/ossf/scorecard/v5/errors"
)

func Test_init_clearsErr(t *testing.T) {
	t.Parallel()
	handler := &checkrunsHandler{errSetup: sce.ErrScorecardInternal}
	handler.init(t.Context(), nil, 0)
	if handler.errSetup != nil {
		t.Errorf("expected nil error, got %v", handler.errSetup)
	}
}

func Test_listCheckRunsForRef_fallsBackForTruncatedGraphQLData(t *testing.T) {
	t.Parallel()
	data := checkRunsGraphQLFixture(t)
	data.Repository.Object.Commit.History.Nodes[0].AssociatedPullRequests.Nodes[0].
		Commits.Nodes[0].Commit.CheckSuites.PageInfo.HasNextPage = true

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/repos/owner/repo/commits/head/check-runs"; got != want {
			t.Errorf("request path = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"total_count": 1,
			"check_runs": [{
				"status": "completed",
				"conclusion": "success",
				"url": "https://api.github.com/check-runs/1",
				"app": {"slug": "github-actions"}
			}]
		}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	client := github.NewClient(server.Client())
	client.BaseURL = baseURL

	handler := new(checkrunsHandler)
	handler.init(t.Context(), &Repo{owner: "owner", repo: "repo"}, 1)
	handler.client = client
	handler.setupOnce.Do(func() {
		handler.checkRunsByRef = parseCheckRuns(data)
	})

	checkRuns, err := handler.listCheckRunsForRef("head")
	if err != nil {
		t.Fatalf("listCheckRunsForRef() error = %v", err)
	}
	if got, want := len(checkRuns), 1; got != want {
		t.Fatalf("len(checkRuns) = %d, want %d", got, want)
	}
	if got, want := checkRuns[0].App.Slug, "github-actions"; got != want {
		t.Errorf("check run app = %q, want %q", got, want)
	}
}

func checkRunsGraphQLFixture(t *testing.T) *checkRunsGraphqlData {
	t.Helper()
	const fixture = `{
		"Repository": {
			"Object": {
				"Commit": {
					"History": {
						"Nodes": [{
							"AssociatedPullRequests": {
								"Nodes": [{
									"HeadRefOid": "head",
									"Commits": {
										"Nodes": [{
											"Commit": {
												"CheckSuites": {
													"Nodes": [{
														"App": {"Slug": "third-party"},
														"Conclusion": "SUCCESS",
														"Status": "COMPLETED"
													}],
													"PageInfo": {"HasNextPage": false}
												}
											}
										}]
									}
								}]
							}
						}]
					}
				}
			}
		}
	}`

	var data checkRunsGraphqlData
	//nolint:musttag // GraphQL response structs are populated by field name and intentionally have no JSON tags.
	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("unmarshal GraphQL fixture: %v", err)
	}
	return &data
}
