// Copyright 2025 OpenSSF Scorecard Authors
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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/shurcooL/githubv4"
)

// RoundTripper that returns canned GraphQL responses in order.
type seqRT struct {
	bodies [][]byte
	i      int
	code   int
}

func (s *seqRT) RoundTrip(req *http.Request) (*http.Response, error) {
	if s.i >= len(s.bodies) {
		return nil, fmt.Errorf("seqRT: no more responses (call #%d)", s.i)
	}
	b := s.bodies[s.i]
	s.i++
	return &http.Response{
		StatusCode: s.code,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(b)),
		Request:    req,
	}, nil
}

func newGraphQLClientWith(jsonBodies ...string) *githubv4.Client {
	bufs := make([][]byte, 0, len(jsonBodies))
	for _, s := range jsonBodies {
		bufs = append(bufs, []byte(s))
	}
	return githubv4.NewClient(&http.Client{Transport: &seqRT{bodies: bufs, code: 200}})
}

func TestTagsHandler_GetTag_TableDriven(t *testing.T) {
	t.Parallel()

	rulesJSONCase1 := `
{
  "data": {
    "repository": {
      "rulesets": {
        "nodes": [
          {
            "name": "Tag PR + checks",
            "enforcement": "ACTIVE",
            "target": "TAG",
            "conditions": { "refName": { "include": ["v1.*"], "exclude": [] } },
            "bypassActors": { "nodes": [
              { "bypassMode": "ALWAYS", "organizationAdmin": true, "repositoryRoleName": "maintainer" }
            ]},
            "rules": { "nodes": [
              {
                "type": "PULL_REQUEST",
                "parameters": {
                  "dismissStaleReviewsOnPush": true,
                  "requireCodeOwnerReview": true,
                  "requireLastPushApproval": true,
                  "requiredApprovingReviewCount": 2,
                  "requiredReviewThreadResolution": true
                }
              },
              {
                "type": "REQUIRED_STATUS_CHECKS",
                "parameters": {
                  "strictRequiredStatusChecksPolicy": true,
                  "requiredStatusChecks": [
                    { "context": "build", "integrationID": 1 },
                    { "context": "test",  "integrationID": 1 }
                  ]
                }
              }
            ]}
          }
        ]
      }
    }
  }
}`

	// Tag refs (must ONLY include fields that `branch` selected via tagsQuery2 -> name)
	tagsJSONCase1 := `
{
  "data": {
    "repository": {
      "refs": {
        "nodes": [
          { "name": "v1.2.3" },
          { "name": "v0.9.0" }
        ]
      }
    }
  }
}`

	// CASE 2 — Another tag with only status checks via a matching ruleset (no PR rule).
	// Strict=false and single context "ci".
	rulesJSONCase2 := `
{
  "data": {
    "repository": {
      "rulesets": {
        "nodes": [
          {
            "name": "Release checks",
            "enforcement": "ACTIVE",
            "target": "TAG",
            "conditions": { "refName": { "include": ["release-*"], "exclude": [] } },
            "bypassActors": { "nodes": [] },
            "rules": { "nodes": [
              {
                "type": "REQUIRED_STATUS_CHECKS",
                "parameters": {
                  "strictRequiredStatusChecksPolicy": false,
                  "requiredStatusChecks": [
                    { "context": "ci", "integrationID": 1 }
                  ]
                }
              }
            ]}
          }
        ]
      }
    }
  }
}`

	tagsJSONCase2 := `
{
  "data": {
    "repository": {
      "refs": {
        "nodes": [
          { "name": "release-2024.10" }
        ]
      }
    }
  }
}`

	// CASE 3 — No ruleset matches tag "v1.9.9"; ≥3 rulesets present; each ruleset has 2 rules.
	rulesJSONCase3 := `
{
  "data": {
    "repository": {
      "rulesets": {
        "nodes": [
          {
            "name": "Only v2 tags",
            "enforcement": "ACTIVE",
            "target": "TAG",
            "conditions": { "refName": { "include": ["v2.*"], "exclude": [] } },
            "bypassActors": { "nodes": [] },
            "rules": { "nodes": [
              {
                "type": "PULL_REQUEST",
                "parameters": {
                  "requiredApprovingReviewCount": 2
                }
              },
              {
                "type": "REQUIRED_STATUS_CHECKS",
                "parameters": {
                  "strictRequiredStatusChecksPolicy": true,
                  "requiredStatusChecks": [
                    { "context": "build", "integrationID": 1 }
                  ]
                }
              }
            ] }
          },
          {
            "name": "Beta rules",
            "enforcement": "ACTIVE",
            "target": "TAG",
            "conditions": { "refName": { "include": ["beta-*"], "exclude": [] } },
            "bypassActors": { "nodes": [] },
            "rules": { "nodes": [
              {
                "type": "PULL_REQUEST",
                "parameters": {
                  "requireCodeOwnerReview": true
                }
              },
              {
                "type": "REQUIRED_STATUS_CHECKS",
                "parameters": {
                  "strictRequiredStatusChecksPolicy": false,
                  "requiredStatusChecks": [
                    { "context": "ci", "integrationID": 1 }
                  ]
                }
              }
            ] }
          },
          {
            "name": "Exclude v1 series",
            "enforcement": "ACTIVE",
            "target": "TAG",
            "conditions": { "refName": { "include": ["*"], "exclude": ["v1.*"] } },
            "bypassActors": { "nodes": [] },
            "rules": { "nodes": [
              {
                "type": "PULL_REQUEST",
                "parameters": {
                  "dismissStaleReviewsOnPush": true
                }
              },
              {
                "type": "REQUIRED_STATUS_CHECKS",
                "parameters": {
                  "strictRequiredStatusChecksPolicy": true,
                  "requiredStatusChecks": [
                    { "context": "lint", "integrationID": 1 }
                  ]
                }
              }
            ] }
          }
        ]
      }
    }
  }
}`

	tagsJSONCase3 := `
{
  "data": {
    "repository": {
      "refs": {
        "nodes": [
          { "name": "v1.9.9" }
        ]
      }
    }
  }
}`

	// CASE 4 — PR-only ruleset applies to "beta-2025.01"; ≥3 rulesets total; matching ruleset has 2 PR rules.
	rulesJSONCase4 := `
{
  "data": {
    "repository": {
      "rulesets": {
        "nodes": [
          {
            "name": "Beta PR gate",
            "enforcement": "ACTIVE",
            "target": "TAG",
            "conditions": { "refName": { "include": ["beta-*"], "exclude": [] } },
            "bypassActors": { "nodes": [] },
            "rules": { "nodes": [
              {
                "type": "PULL_REQUEST",
                "parameters": {
                  "requiredApprovingReviewCount": 1
                }
              },
              {
                "type": "PULL_REQUEST",
                "parameters": {
                  "dismissStaleReviewsOnPush": true
                }
              }
            ] }
          },
          {
            "name": "General checks for v2",
            "enforcement": "ACTIVE",
            "target": "TAG",
            "conditions": { "refName": { "include": ["v2.*"], "exclude": [] } },
            "bypassActors": { "nodes": [] },
            "rules": { "nodes": [
              {
                "type": "REQUIRED_STATUS_CHECKS",
                "parameters": {
                  "strictRequiredStatusChecksPolicy": true,
                  "requiredStatusChecks": [
                    { "context": "unit", "integrationID": 1 }
                  ]
                }
              },
              {
                "type": "REQUIRED_STATUS_CHECKS",
                "parameters": {
                  "strictRequiredStatusChecksPolicy": true,
                  "requiredStatusChecks": [
                    { "context": "integration", "integrationID": 1 }
                  ]
                }
              }
            ] }
          },
          {
            "name": "Exclude betas",
            "enforcement": "ACTIVE",
            "target": "TAG",
            "conditions": { "refName": { "include": ["*"], "exclude": ["beta-*"] } },
            "bypassActors": { "nodes": [] },
            "rules": { "nodes": [
              {
                "type": "PULL_REQUEST",
                "parameters": {
                  "requireCodeOwnerReview": true
                }
              },
              {
                "type": "REQUIRED_STATUS_CHECKS",
                "parameters": {
                  "strictRequiredStatusChecksPolicy": true,
                  "requiredStatusChecks": [
                    { "context": "lint", "integrationID": 1 }
                  ]
                }
              }
            ] }
          }
        ]
      }
    }
  }
}`

	tagsJSONCase4 := `
{
  "data": {
    "repository": {
      "refs": {
        "nodes": [
          { "name": "beta-2025.01" }
        ]
      }
    }
  }
}`

	tests := []struct {
		wantCodeOwners                    *bool
		wantRequireLinearHistory          *bool
		wantStatusChecksStrictUpToDate    *bool
		wantStatusChecksRequired          *bool
		wantRequireLastPushApproval       *bool
		wantEnforceAdmins                 *bool
		wantDismissStale                  *bool
		wantPRApprovals                   *int32
		wantAllowDeletions                *bool
		wantAllowForcePushes              *bool
		wantPRRequired                    *bool
		tag                               string
		rulesJSON                         string
		wantName                          string
		name                              string
		repo                              string
		owner                             string
		tagsJSON                          string
		wantStatusChecksContextsUnordered []string
		wantProtected                     bool
	}{
		{
			name:      "ruleset-applies-v1.2.3",
			rulesJSON: rulesJSONCase1,
			tagsJSON:  tagsJSONCase1,
			owner:     "o",
			repo:      "r",
			tag:       "v1.2.3",

			wantName:      "v1.2.3",
			wantProtected: true,
			// No BPR/RUR in tags JSON -> these toggles remain nil
			wantEnforceAdmins:        nil,
			wantAllowDeletions:       nil,
			wantAllowForcePushes:     nil,
			wantRequireLinearHistory: nil,

			// From ruleset:
			wantPRRequired:                    boolPtr(true),
			wantPRApprovals:                   int32Ptr(2),
			wantDismissStale:                  boolPtr(true),
			wantCodeOwners:                    boolPtr(true),
			wantRequireLastPushApproval:       boolPtr(true),
			wantStatusChecksRequired:          boolPtr(true),
			wantStatusChecksStrictUpToDate:    boolPtr(true),
			wantStatusChecksContextsUnordered: []string{"build", "test"},
		},
		{
			name:      "release-tag-status-checks-only",
			rulesJSON: rulesJSONCase2,
			tagsJSON:  tagsJSONCase2,
			owner:     "o",
			repo:      "r",
			tag:       "release-2024.10",

			wantName:                 "release-2024.10",
			wantProtected:            true,
			wantEnforceAdmins:        nil,
			wantAllowDeletions:       nil,
			wantAllowForcePushes:     nil,
			wantRequireLinearHistory: nil,

			// From ruleset (no PR rules):
			wantPRRequired:                    nil,
			wantPRApprovals:                   nil,
			wantDismissStale:                  nil,
			wantCodeOwners:                    nil,
			wantRequireLastPushApproval:       nil,
			wantStatusChecksRequired:          boolPtr(true),
			wantStatusChecksStrictUpToDate:    boolPtr(false),
			wantStatusChecksContextsUnordered: []string{"ci"},
		},
		{
			name:      "no-ruleset-match-unprotected",
			rulesJSON: rulesJSONCase3,
			tagsJSON:  tagsJSONCase3,
			owner:     "o",
			repo:      "r",
			tag:       "v1.9.9",

			wantName:                 "v1.9.9",
			wantProtected:            false, // none of the 3 rulesets apply
			wantEnforceAdmins:        nil,
			wantAllowDeletions:       nil,
			wantAllowForcePushes:     nil,
			wantRequireLinearHistory: nil,

			wantPRRequired:                    nil,
			wantPRApprovals:                   nil,
			wantDismissStale:                  nil,
			wantCodeOwners:                    nil,
			wantRequireLastPushApproval:       nil,
			wantStatusChecksRequired:          nil,
			wantStatusChecksStrictUpToDate:    nil,
			wantStatusChecksContextsUnordered: nil,
		},
		{
			name:      "pr-only-two-rules-beta",
			rulesJSON: rulesJSONCase4,
			tagsJSON:  tagsJSONCase4,
			owner:     "o",
			repo:      "r",
			tag:       "beta-2025.01",

			wantName:                 "beta-2025.01",
			wantProtected:            true,
			wantEnforceAdmins:        nil,
			wantAllowDeletions:       nil,
			wantAllowForcePushes:     nil,
			wantRequireLinearHistory: nil,

			// From matching ruleset's two PR rules:
			wantPRRequired:              boolPtr(true),
			wantPRApprovals:             nil,
			wantDismissStale:            boolPtr(true),
			wantCodeOwners:              nil,
			wantRequireLastPushApproval: nil,

			wantStatusChecksRequired:          nil,
			wantStatusChecksStrictUpToDate:    nil,
			wantStatusChecksContextsUnordered: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			graph := newGraphQLClientWith(tc.rulesJSON, tc.tagsJSON)

			var h tagsHandler
			h.graphClient = graph
			ctx := context.Background()
			h.init(ctx, &Repo{owner: tc.owner, repo: tc.repo})

			if err := h.setup(); err != nil {
				t.Fatalf("setup() error: %v", err)
			}

			got, err := h.getTag(tc.tag)
			if err != nil {
				t.Fatalf("getTag(%q) error: %v", tc.tag, err)
			}
			if got == nil {
				t.Fatalf("nil RepoRef")
			}

			// Name / Protected
			if got.Name == nil || *got.Name != tc.wantName {
				t.Fatalf("Name = %v, want %q", derefStr(got.Name), tc.wantName)
			}
			if got.Protected == nil || *got.Protected != tc.wantProtected {
				t.Fatalf("Protected = %v, want %v", derefBool(got.Protected), tc.wantProtected)
			}

			// Low-level toggles (should stay nil without BPR/RUR in the JSON)
			assertOptBool(t, "EnforceAdmins", got.ProtectionRule.EnforceAdmins, tc.wantEnforceAdmins)
			assertOptBool(t, "AllowDeletions", got.ProtectionRule.AllowDeletions, tc.wantAllowDeletions)
			assertOptBool(t, "AllowForcePushes", got.ProtectionRule.AllowForcePushes, tc.wantAllowForcePushes)
			assertOptBool(t, "RequireLinearHistory", got.ProtectionRule.RequireLinearHistory, tc.wantRequireLinearHistory)

			// PR Rule expectations
			assertOptBool(t, "PR.Required", got.ProtectionRule.PullRequestRule.Required, tc.wantPRRequired)
			assertOptInt32(t, "PR.RequiredApprovingReviewCount", got.ProtectionRule.PullRequestRule.RequiredApprovingReviewCount, tc.wantPRApprovals)
			assertOptBool(t, "PR.DismissStaleReviews", got.ProtectionRule.PullRequestRule.DismissStaleReviews, tc.wantDismissStale)
			assertOptBool(t, "PR.RequireCodeOwnerReviews", got.ProtectionRule.PullRequestRule.RequireCodeOwnerReviews, tc.wantCodeOwners)
			assertOptBool(t, "RequireLastPushApproval", got.ProtectionRule.RequireLastPushApproval, tc.wantRequireLastPushApproval)

			// Status checks expectations
			assertOptBool(t, "Checks.RequiresStatusChecks", got.ProtectionRule.CheckRules.RequiresStatusChecks, tc.wantStatusChecksRequired)
			assertOptBool(t, "Checks.UpToDateBeforeMerge", got.ProtectionRule.CheckRules.UpToDateBeforeMerge, tc.wantStatusChecksStrictUpToDate)

			// Contexts (order-insensitive)
			if tc.wantStatusChecksContextsUnordered != nil {
				gotCtx := append([]string(nil), got.ProtectionRule.CheckRules.Contexts...)
				want := set(tc.wantStatusChecksContextsUnordered...)
				if !sameSet(gotCtx, want) {
					t.Fatalf("Checks.Contexts = %v, want set %v", gotCtx, tc.wantStatusChecksContextsUnordered)
				}
			}
		})
	}
}

/**********************
 * Small helpers
 **********************/

func boolPtr(b bool) *bool    { return &b }
func int32Ptr(i int32) *int32 { return &i }

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func assertOptBool(t *testing.T, field string, got, want *bool) {
	t.Helper()
	// If we don't care (want is nil), don't assert.
	if want == nil {
		return
	}
	if got == nil {
		t.Fatalf("%s = <nil>, want %v", field, *want)
	}
	if *got != *want {
		t.Fatalf("%s = %v, want %v", field, *got, *want)
	}
}

func assertOptInt32(t *testing.T, field string, got, want *int32) {
	t.Helper()
	// If we don't care (want is nil), don't assert.
	if want == nil {
		return
	}
	if got == nil {
		t.Fatalf("%s = <nil>, want %v", field, *want)
	}
	if *got != *want {
		t.Fatalf("%s = %v, want %v", field, *got, *want)
	}
}

func set(ss ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func sameSet(got []string, want map[string]struct{}) bool {
	if len(got) != len(want) {
		return false
	}
	for _, s := range got {
		if _, ok := want[s]; !ok {
			return false
		}
	}
	return true
}
