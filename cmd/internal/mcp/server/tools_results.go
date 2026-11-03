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

package server

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/sync/errgroup"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/catalog"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/model"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/provider"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/scorecardref"
)

// maxCompareRepos bounds the number of repositories compare_repos will process.
const maxCompareRepos = 10

// compareConcurrency bounds how many repositories compare_repos fetches at once.
// It is kept modest to stay a good citizen toward the upstream Scorecard API.
const compareConcurrency = 5

// commitPattern advertises the 40-char hex SHA format that
// scorecardref.ValidateCommit enforces. The empty alternation keeps an
// explicitly-empty (i.e. "no commit") value valid, matching ValidateCommit,
// since commit is an optional field.
const commitPattern = `^([0-9a-fA-F]{40})?$`

// inputSchemaFor infers the JSON Schema for a tool input type (identical to what
// the SDK would infer) and applies patch. It exists so we can advertise
// constraints the jsonschema struct tag cannot express in this SDK, where the
// tag is description-only. These are client-side, fail-fast hints; the tool
// handlers remain the authoritative validators (scorecardref.ValidateCommit and
// the maxCompareRepos truncation path).
func inputSchemaFor[T any](patch func(*jsonschema.Schema)) *jsonschema.Schema {
	s, err := jsonschema.For[T](nil)
	if err != nil {
		// The input types are fixed and trivially schema-able; a failure here is
		// a programming error, so fail loudly at registration as mcp.AddTool does.
		panic(fmt.Sprintf("build input schema: %v", err))
	}
	patch(s)
	return s
}

// Bounds for get_check_result's max_details input. An omitted (zero) value
// falls back to defaultMaxDetails; any provided value is clamped to
// [minMaxDetails, maxMaxDetails]. Verified against live data before choosing the
// default: Pinned-Dependencies details ranged 5-186 across sampled repos and
// Token-Permissions hit 377 on envoyproxy/envoy, so truncation must be generic.
const (
	defaultMaxDetails = 50
	minMaxDetails     = 1
	maxMaxDetails     = 500
)

// --- get_repo_score ---

type getRepoScoreInput struct {
	Repo   string `json:"repo" jsonschema:"platform/owner/repo (platform optional; github.com or gitlab.com)"`
	Commit string `json:"commit,omitempty" jsonschema:"optional 40-char hex commit SHA to pin results"`
}

type checkSummary struct {
	Name   string `json:"name"`
	Reason string `json:"reason,omitempty"`
	Score  int    `json:"score" jsonschema:"score 0-10, or -1 for inconclusive"`
}

type repoScoreOutput struct {
	Attribution   *model.Attribution  `json:"attribution,omitempty"`
	Repo          model.RepoRef       `json:"repo"`
	Scorecard     model.ScorecardInfo `json:"scorecard"`
	Commit        string              `json:"commit,omitempty" jsonschema:"resolved commit SHA of the results"`
	Date          string              `json:"date,omitempty" jsonschema:"when the scan was generated (RFC3339)"`
	Source        model.Source        `json:"source" jsonschema:"which provider produced the result"`
	Checks        []checkSummary      `json:"checks"`
	Caveats       []string            `json:"caveats,omitempty"`
	MissingChecks []string            `json:"missing_checks,omitempty" jsonschema:"expected checks the result lacks"`
	Score         float64             `json:"score" jsonschema:"aggregate score 0-10, or -1 for inconclusive"`
	Complete      bool                `json:"complete" jsonschema:"whether the result has every check for its platform"`
}

// --- get_check_result ---

type getCheckResultInput struct {
	Repo       string `json:"repo" jsonschema:"platform/owner/repo (platform optional)"`
	Check      string `json:"check" jsonschema:"check name, e.g. Branch-Protection (see list_checks)"`
	Commit     string `json:"commit,omitempty" jsonschema:"optional 40-char hex commit SHA"`
	MaxDetails int    `json:"max_details,omitempty" jsonschema:"max detail lines to return; default 50, clamped 1-500"`
}

// checkDetail is get_check_result's per-check output. Unlike the shared
// model.Check, it caps Details at the caller's max_details and reports the
// pre-truncation count so a client knows the payload was shortened. Truncation
// is applied here in the handler, never in the provider, so Provider.GetResult
// keeps returning the fullest result it can (preserving darnit's full-JSON
// contract).
type checkDetail struct {
	Documentation    *model.Documentation `json:"documentation,omitempty"`
	Name             string               `json:"name"`
	Reason           string               `json:"reason,omitempty"`
	Details          []string             `json:"details,omitempty"`
	Score            int                  `json:"score" jsonschema:"score from 0-10, or -1 for inconclusive"`
	DetailsTotal     int                  `json:"details_total" jsonschema:"number of detail lines before truncation"`
	DetailsTruncated bool                 `json:"details_truncated" jsonschema:"true if details was capped by max_details"`
}

type getCheckResultOutput struct {
	Attribution *model.Attribution  `json:"attribution,omitempty"`
	Repo        model.RepoRef       `json:"repo"`
	Scorecard   model.ScorecardInfo `json:"scorecard"`
	Commit      string              `json:"commit,omitempty"`
	Date        string              `json:"date,omitempty"`
	Source      model.Source        `json:"source"`
	Caveats     []string            `json:"caveats,omitempty"`
	Check       checkDetail         `json:"check"`
}

// --- compare_repos ---

type compareReposInput struct {
	Repos []string `json:"repos" jsonschema:"repository references to compare, each platform/owner/repo"`
}

type repoComparison struct {
	Input  string         `json:"input"`
	Repo   *model.RepoRef `json:"repo,omitempty"`
	Score  *float64       `json:"score,omitempty" jsonschema:"aggregate score 0-10, or -1 for inconclusive"`
	Date   string         `json:"date,omitempty"`
	Source model.Source   `json:"source,omitempty"`
	Error  string         `json:"error,omitempty" jsonschema:"set when this repo could not be retrieved"`
}

type compareReposOutput struct {
	Note      string           `json:"note,omitempty"`
	Results   []repoComparison `json:"results"`
	Truncated bool             `json:"truncated" jsonschema:"true if the input list was truncated to the maximum"`
}

// registerResultTools registers the scorecard-results capability tools.
func registerResultTools(s *mcp.Server, p provider.Provider, cat *catalog.Catalog) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "get_repo_score",
		Description: "Return an OpenSSF Scorecard aggregate score and a per-check summary " +
			"(name, score, short reason) for one repository. This is a compact summary; use " +
			"get_check_result for a single check's full details, and explain_check for a " +
			"check's methodology. Data comes from the OpenSSF Scorecard REST API " +
			"(https://api.scorecard.dev).",
		Annotations: readOnlyAnnotations("Get repository Scorecard score", true),
		InputSchema: inputSchemaFor[getRepoScoreInput](func(s *jsonschema.Schema) {
			s.Properties["commit"].Pattern = commitPattern
		}),
	}, func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		in getRepoScoreInput,
	) (*mcp.CallToolResult, repoScoreOutput, error) {
		ref, err := scorecardref.Parse(in.Repo)
		if err != nil {
			return nil, repoScoreOutput{}, fmt.Errorf("repo: %w", err)
		}
		if err := scorecardref.ValidateCommit(in.Commit); err != nil {
			return nil, repoScoreOutput{}, fmt.Errorf("commit: %w", err)
		}
		res, err := fetchResult(ctx, p, cat, ref, in.Commit)
		if err != nil {
			return nil, repoScoreOutput{}, err
		}
		return nil, toRepoScore(res), nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "get_check_result",
		Description: "Return the result (score, reason, details) for a single named OpenSSF " +
			"Scorecard check on one repository. Detail lines are capped (default 50; adjust with " +
			"max_details), and details_truncated and details_total report any truncation. Use " +
			"list_checks to discover valid check names and explain_check for a check's methodology " +
			"and remediation. Data comes from the OpenSSF Scorecard REST API.",
		Annotations: readOnlyAnnotations("Get one Scorecard check result", true),
		InputSchema: inputSchemaFor[getCheckResultInput](func(s *jsonschema.Schema) {
			s.Properties["commit"].Pattern = commitPattern
		}),
	}, func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		in getCheckResultInput,
	) (*mcp.CallToolResult, getCheckResultOutput, error) {
		ref, err := scorecardref.Parse(in.Repo)
		if err != nil {
			return nil, getCheckResultOutput{}, fmt.Errorf("repo: %w", err)
		}
		if err := scorecardref.ValidateCommit(in.Commit); err != nil {
			return nil, getCheckResultOutput{}, fmt.Errorf("commit: %w", err)
		}
		if in.Check == "" {
			return nil, getCheckResultOutput{}, errCheckRequired
		}
		res, err := fetchResult(ctx, p, cat, ref, in.Commit)
		if err != nil {
			return nil, getCheckResultOutput{}, err
		}
		c, ok := res.FindCheck(in.Check)
		if !ok {
			return nil, getCheckResultOutput{}, checkAbsenceError(in.Check, ref, cat)
		}
		return nil, getCheckResultOutput{
			Repo:        res.Repo,
			Commit:      res.Commit,
			Date:        res.Date,
			Scorecard:   res.Scorecard,
			Source:      res.Source,
			Check:       toCheckDetail(c, clampMaxDetails(in.MaxDetails)),
			Caveats:     res.Caveats,
			Attribution: res.Attribution,
		}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "compare_repos",
		Description: "Return the OpenSSF Scorecard aggregate scores of several repositories for " +
			"comparison. Each repository's score and provenance are returned; a per-repository " +
			"failure is reported inline rather than failing the whole call. If more than the " +
			"supported maximum are supplied, the list is truncated. Data comes from the OpenSSF " +
			"Scorecard REST API.",
		Annotations: readOnlyAnnotations("Compare repository Scorecard scores", true),
		InputSchema: inputSchemaFor[compareReposInput](func(s *jsonschema.Schema) {
			// Advertise the cap as a client-side hint only; the handler still
			// truncates (and reports it) so a client that ignores this hint or
			// does not validate is handled gracefully rather than rejected.
			s.Properties["repos"].Description += fmt.Sprintf(
				"; at most %d are compared, any beyond that are dropped and reported via the truncated flag",
				maxCompareRepos)
		}),
	}, func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		in compareReposInput,
	) (*mcp.CallToolResult, compareReposOutput, error) {
		if len(in.Repos) == 0 {
			return nil, compareReposOutput{}, errReposRequired
		}

		var out compareReposOutput
		repos := in.Repos
		if len(repos) > maxCompareRepos {
			out.Truncated = true
			out.Note = fmt.Sprintf("only the first %d of %d repositories were compared", maxCompareRepos, len(in.Repos))
			repos = repos[:maxCompareRepos]
		}

		// Fetch concurrently but bounded, so we stay a good citizen toward the
		// upstream API. Each result is written to its own slot to preserve input
		// order; per-repository failures are captured inline by compareOne, so
		// the errgroup goroutines never return an error.
		results := make([]repoComparison, len(repos))
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(compareConcurrency)
		for i, r := range repos {
			g.Go(func() error {
				results[i] = compareOne(gctx, p, r)
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			// Unreachable in practice: compareOne captures every per-repository
			// failure inline, so the goroutines always return nil. Surfaced
			// defensively rather than discarded.
			return nil, compareReposOutput{}, fmt.Errorf("comparing repositories: %w", err)
		}
		out.Results = results
		return nil, out, nil
	})
}

// fetchResult retrieves a repository's result and records which expected
// checks it lacks.
func fetchResult(
	ctx context.Context,
	p provider.Provider,
	cat *catalog.Catalog,
	ref model.RepoRef,
	commit string,
) (*model.Result, error) {
	res, err := p.GetResult(ctx, ref, commit)
	if err != nil {
		return nil, fmt.Errorf("fetching results: %w", err)
	}
	annotateCoverage(res, cat)
	return res, nil
}

// annotateCoverage lists the non-experimental checks for the result's
// platform that the result does not include, marks the result complete only
// when there are none, and adds a caveat naming them.
func annotateCoverage(res *model.Result, cat *catalog.Catalog) {
	var missing []string
	for _, name := range cat.ExpectedChecks(res.Repo.Platform) {
		if _, ok := res.FindCheck(name); !ok {
			missing = append(missing, name)
		}
	}
	res.MissingChecks = missing
	res.Complete = len(missing) == 0
	if !res.Complete {
		res.Caveats = append(slices.Clip(res.Caveats), fmt.Sprintf(
			"This result does not include the %s checks; `scorecard --repo=%s` computes every check locally.",
			strings.Join(missing, ", "), res.Repo.String()))
	}
}

// checkAbsenceError explains why a requested check is missing from a
// repository's result, distinguishing an unknown check name, a known but
// experimental check, a check that does not run on the repository's
// platform, and a check the result does not include.
func checkAbsenceError(name string, ref model.RepoRef, cat *catalog.Catalog) error {
	detail, err := cat.Explain(name)
	if err != nil {
		return fmt.Errorf("%w: %q; use list_checks to see valid check names", errCheckUnknown, name)
	}
	if detail.Experimental {
		return fmt.Errorf(
			"%w: %q is only included when a scan runs with SCORECARD_EXPERIMENTAL set",
			errCheckExperimental, name)
	}
	if !slices.Contains(cat.ExpectedChecks(ref.Platform), detail.Name) {
		return fmt.Errorf("%w: %q does not run on %s", errCheckUnsupported, detail.Name, ref.Platform)
	}
	return fmt.Errorf(
		"%w: the result for %s does not include %q; `scorecard --repo=%s --checks=%s` computes it locally",
		errCheckMissing, ref.String(), detail.Name, ref.String(), detail.Name)
}

// compareOne fetches one repository's aggregate score for compare_repos,
// reporting any per-repository failure inline.
func compareOne(ctx context.Context, p provider.Provider, r string) repoComparison {
	entry := repoComparison{Input: r}
	ref, err := scorecardref.Parse(r)
	if err != nil {
		entry.Error = err.Error()
		return entry
	}
	entry.Repo = &ref
	res, err := p.GetResult(ctx, ref, "")
	if err != nil {
		entry.Error = err.Error()
		return entry
	}
	score := res.Score
	entry.Score = &score
	entry.Date = res.Date
	entry.Source = res.Source
	return entry
}

// clampMaxDetails resolves the max_details input: an omitted (zero) value uses
// the default, and any provided value is clamped to [minMaxDetails, maxMaxDetails].
func clampMaxDetails(n int) int {
	if n == 0 {
		return defaultMaxDetails
	}
	if n < minMaxDetails {
		return minMaxDetails
	}
	if n > maxMaxDetails {
		return maxMaxDetails
	}
	return n
}

// toCheckDetail converts a model.Check into get_check_result's output, capping
// Details at maxDetails and recording the pre-truncation total.
func toCheckDetail(c model.Check, maxDetails int) checkDetail {
	total := len(c.Details)
	details := c.Details
	truncated := false
	if total > maxDetails {
		details = details[:maxDetails]
		truncated = true
	}
	return checkDetail{
		Name:             c.Name,
		Score:            c.Score,
		Reason:           c.Reason,
		Details:          details,
		DetailsTotal:     total,
		DetailsTruncated: truncated,
		Documentation:    c.Documentation,
	}
}

func toRepoScore(res *model.Result) repoScoreOutput {
	summaries := make([]checkSummary, 0, len(res.Checks))
	for _, c := range res.Checks {
		summaries = append(summaries, checkSummary{Name: c.Name, Score: c.Score, Reason: c.Reason})
	}
	return repoScoreOutput{
		Repo:          res.Repo,
		Commit:        res.Commit,
		Date:          res.Date,
		Scorecard:     res.Scorecard,
		Source:        res.Source,
		Score:         res.Score,
		Checks:        summaries,
		Caveats:       res.Caveats,
		Attribution:   res.Attribution,
		MissingChecks: res.MissingChecks,
		Complete:      res.Complete,
	}
}
