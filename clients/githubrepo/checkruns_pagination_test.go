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

package githubrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/google/go-github/v82/github"
)

func TestParseCheckRunsSkipsIncompleteSuites(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		hasNextPage bool
		wantCached  bool
	}{
		{name: "complete page", wantCached: true},
		{name: "truncated page", hasNextPage: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			suites := make([]map[string]any, checksToAnalyze)
			for i := range suites {
				suites[i] = map[string]any{
					"App":        map[string]any{"Slug": "other-check"},
					"Status":     "COMPLETED",
					"Conclusion": "SUCCESS",
				}
			}
			data := map[string]any{
				"Repository": map[string]any{
					"Object": map[string]any{
						"Commit": map[string]any{
							"History": map[string]any{
								"Nodes": []any{map[string]any{
									"AssociatedPullRequests": map[string]any{
										"Nodes": []any{map[string]any{
											"HeadRefOid": "head-sha",
											"Commits": map[string]any{
												"Nodes": []any{map[string]any{
													"Commit": map[string]any{
														"CheckSuites": map[string]any{
															"PageInfo": map[string]any{"HasNextPage": tc.hasNextPage},
															"Nodes":    suites,
														},
													},
												}},
											},
										}},
									},
								}},
							},
						},
					},
				},
			}
			payload, err := json.Marshal(data)
			if err != nil {
				t.Fatal(err)
			}
			var response checkRunsGraphqlData
			if err := json.Unmarshal(payload, &response); err != nil {
				t.Fatal(err)
			}

			cached := parseCheckRuns(&response)
			got, ok := cached["head-sha"]
			if ok != tc.wantCached {
				t.Fatalf("cached=%t, want %t", ok, tc.wantCached)
			}
			if tc.wantCached && len(got) != checksToAnalyze {
				t.Fatalf("cached %d suites, want %d", len(got), checksToAnalyze)
			}
		})
	}
}

func TestListCheckRunsForRefPaginatesAndCaches(t *testing.T) {
	t.Parallel()

	const totalRuns = 203
	requests := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/ossf/scorecard/commits/head-sha/check-runs" {
			t.Errorf("unexpected path %q", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("per_page") != "100" {
			t.Errorf("per_page=%q, want 100", r.URL.Query().Get("per_page"))
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		requests++
		switch page {
		case 1, 2, 3:
		default:
			t.Errorf("unexpected page %d", page)
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}
		count := 100
		if page == 3 {
			count = 3
		}
		runs := make([]map[string]any, count)
		for i := range runs {
			slug := "unrelated-integration"
			if page == 3 && i == 2 {
				slug = "github-actions"
			}
			runs[i] = map[string]any{
				"status": "completed", "conclusion": "success",
				"app": map[string]any{"slug": slug},
			}
		}
		if page < 3 {
			w.Header().Set("Link", fmt.Sprintf("<%s%s?page=%d&per_page=100>; rel=\"next\"", server.URL, r.URL.Path, page+1))
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"total_count": totalRuns,
			"check_runs":  runs,
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := github.NewClient(server.Client())
	var err error
	client.BaseURL, err = url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	var handler checkrunsHandler
	handler.init(context.Background(), &Repo{owner: "ossf", repo: "scorecard"}, 1)
	handler.client = client
	// Mark the GraphQL setup as complete to exercise the REST fallback.
	handler.setupOnce.Do(func() {})

	runs, err := handler.listCheckRunsForRef("head-sha")
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != totalRuns {
		t.Fatalf("got %d check runs, want %d", len(runs), totalRuns)
	}
	if runs[len(runs)-1].App.Slug != "github-actions" {
		t.Fatalf("last check run app=%q, want github-actions", runs[len(runs)-1].App.Slug)
	}
	if requests != 3 {
		t.Fatalf("requests=%d, want 3", requests)
	}
	cached, err := handler.listCheckRunsForRef("head-sha")
	if err != nil {
		t.Fatal(err)
	}
	if len(cached) != totalRuns || requests != 3 {
		t.Fatalf("cache miss: got %d runs and %d HTTP requests", len(cached), requests)
	}
}
