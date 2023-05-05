// Copyright 2021 OpenSSF Scorecard Authors
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

package stats

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	// RemainingTokens measures the remaining number of API tokens.
	RemainingTokens = stats.Int64("RemainingTokens",
		"Measures the remaining count of API tokens", stats.UnitDimensionless)
	// RetryAfter measures the retry delay when dealing with secondary rate limits.
	RetryAfter = stats.Int64("RetryAfter",
		"Measures the retry delay when dealing with secondary rate limits", stats.UnitSeconds)
	// TokenIndex is the tag key for specifying a unique token.
	TokenIndex = tag.MustNewKey("tokenIndex")
	// ResourceType specifies the type of GitHub resource.
	ResourceType = tag.MustNewKey("resourceType")

	// GithubTokens tracks the usage/remaining stats per token per resource-type.
	GithubTokens = view.View{
		Name:        "GithubTokens",
		Description: "Token usage/remaining stats for Github API tokens",
		Measure:     RemainingTokens,
		TagKeys:     []tag.Key{TokenIndex, ResourceType},
		Aggregation: view.LastValue(),
	}
)
