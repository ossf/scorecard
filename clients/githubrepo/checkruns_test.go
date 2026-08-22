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
	"testing"

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

func Test_parseCheckRuns_skipsTruncatedCheckSuites(t *testing.T) {
	t.Parallel()
	const headSHA = "98deba2249aaf487cd8b609939f75b91a54b5bd7"
	data := checkRunsGraphqlFixture(t, headSHA, true)
	got := parseCheckRuns(data)
	if _, cached := got[headSHA]; cached {
		t.Fatalf("truncated checkSuites page was cached for %s", headSHA)
	}
}

func Test_parseCheckRuns_cachesCompleteCheckSuites(t *testing.T) {
	t.Parallel()
	const headSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	data := checkRunsGraphqlFixture(t, headSHA, false)
	got := parseCheckRuns(data)
	crs, cached := got[headSHA]
	if !cached {
		t.Fatalf("complete checkSuites page was not cached for %s", headSHA)
	}
	if len(crs) != 1 {
		t.Fatalf("got %d check runs, want 1", len(crs))
	}
	if crs[0].App.Slug != "github-actions" {
		t.Fatalf("got app slug %q, want github-actions", crs[0].App.Slug)
	}
	if crs[0].Status != "completed" || crs[0].Conclusion != "success" {
		t.Fatalf("got status=%q conclusion=%q", crs[0].Status, crs[0].Conclusion)
	}
}

func checkRunsGraphqlFixture(t *testing.T, headSHA string, hasNextPage bool) *checkRunsGraphqlData {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"Repository": map[string]any{
			"Object": map[string]any{
				"Commit": map[string]any{
					"History": map[string]any{
						"Nodes": []any{
							map[string]any{
								"AssociatedPullRequests": map[string]any{
									"Nodes": []any{
										map[string]any{
											"HeadRefOid": headSHA,
											"Commits": map[string]any{
												"Nodes": []any{
													map[string]any{
														"Commit": map[string]any{
															"CheckSuites": map[string]any{
																"Nodes": []any{
																	map[string]any{
																		"App":        map[string]any{"Slug": "github-actions"},
																		"Conclusion": "SUCCESS",
																		"Status":     "COMPLETED",
																	},
																},
																"PageInfo": map[string]any{
																	"HasNextPage": hasNextPage,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var data checkRunsGraphqlData
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return &data
}
