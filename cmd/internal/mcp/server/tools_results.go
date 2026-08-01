package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/catalog"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/model"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/provider"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/scorecardref"
)

// maxCompareRepos bounds the number of repositories compare_repos will process.
const maxCompareRepos = 10

// --- get_repo_score ---

type getRepoScoreInput struct {
	Repo   string `json:"repo" jsonschema:"platform/owner/repo (platform optional; github.com or gitlab.com)"`
	Commit string `json:"commit,omitempty" jsonschema:"optional 40-char hex commit SHA to pin results"`
}

type checkSummary struct {
	Name   string `json:"name"`
	Score  int    `json:"score" jsonschema:"score 0-10, or -1 for inconclusive"`
	Reason string `json:"reason,omitempty"`
}

type repoScoreOutput struct {
	Repo        model.RepoRef       `json:"repo"`
	Commit      string              `json:"commit,omitempty" jsonschema:"resolved commit SHA of the results"`
	Date        string              `json:"date,omitempty" jsonschema:"when the scan was generated (RFC3339)"`
	Scorecard   model.ScorecardInfo `json:"scorecard"`
	Source      model.Source        `json:"source" jsonschema:"which provider produced the result"`
	Score       float64             `json:"score" jsonschema:"aggregate score 0-10, or -1 for inconclusive"`
	Checks      []checkSummary      `json:"checks"`
	Caveats     []string            `json:"caveats,omitempty"`
	Attribution *model.Attribution  `json:"attribution,omitempty"`
	Complete    bool                `json:"complete" jsonschema:"whether all checks are present (false for cached)"`
}

// --- get_check_result ---

type getCheckResultInput struct {
	Repo   string `json:"repo" jsonschema:"platform/owner/repo (platform optional)"`
	Check  string `json:"check" jsonschema:"check name, e.g. Branch-Protection (see list_checks)"`
	Commit string `json:"commit,omitempty" jsonschema:"optional 40-char hex commit SHA"`
}

type getCheckResultOutput struct {
	Repo        model.RepoRef       `json:"repo"`
	Commit      string              `json:"commit,omitempty"`
	Date        string              `json:"date,omitempty"`
	Scorecard   model.ScorecardInfo `json:"scorecard"`
	Source      model.Source        `json:"source"`
	Check       model.Check         `json:"check"`
	Caveats     []string            `json:"caveats,omitempty"`
	Attribution *model.Attribution  `json:"attribution,omitempty"`
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
	Results   []repoComparison `json:"results"`
	Truncated bool             `json:"truncated" jsonschema:"true if the input list was truncated to the maximum"`
	Note      string           `json:"note,omitempty"`
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
	}, func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		in getRepoScoreInput,
	) (*mcp.CallToolResult, repoScoreOutput, error) {
		ref, err := scorecardref.Parse(in.Repo)
		if err != nil {
			return nil, repoScoreOutput{}, err
		}
		if err := scorecardref.ValidateCommit(in.Commit); err != nil {
			return nil, repoScoreOutput{}, err
		}
		res, err := p.GetResult(ctx, ref, in.Commit)
		if err != nil {
			return nil, repoScoreOutput{}, err
		}
		return nil, toRepoScore(res), nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "get_check_result",
		Description: "Return the full result (score, reason, details) for a single named OpenSSF " +
			"Scorecard check on one repository. Use list_checks to discover valid check names and " +
			"explain_check for a check's methodology and remediation. Data comes from the OpenSSF " +
			"Scorecard REST API.",
		Annotations: readOnlyAnnotations("Get one Scorecard check result", true),
	}, func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		in getCheckResultInput,
	) (*mcp.CallToolResult, getCheckResultOutput, error) {
		ref, err := scorecardref.Parse(in.Repo)
		if err != nil {
			return nil, getCheckResultOutput{}, err
		}
		if err := scorecardref.ValidateCommit(in.Commit); err != nil {
			return nil, getCheckResultOutput{}, err
		}
		if in.Check == "" {
			return nil, getCheckResultOutput{}, errCheckRequired
		}
		res, err := p.GetResult(ctx, ref, in.Commit)
		if err != nil {
			return nil, getCheckResultOutput{}, err
		}
		c, ok := res.FindCheck(in.Check)
		if !ok {
			return nil, getCheckResultOutput{}, checkAbsenceError(in.Check, ref.String(), cat, p)
		}
		return nil, getCheckResultOutput{
			Repo:        res.Repo,
			Commit:      res.Commit,
			Date:        res.Date,
			Scorecard:   res.Scorecard,
			Source:      res.Source,
			Check:       c,
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

		for _, r := range repos {
			out.Results = append(out.Results, compareOne(ctx, p, r))
		}
		return nil, out, nil
	})
}

// checkAbsenceError explains why a requested check is missing from a
// repository's results, distinguishing an unknown check name, a known but
// experimental check, and a known check this provider's scan omits.
func checkAbsenceError(name, repo string, cat *catalog.Catalog, p provider.Provider) error {
	detail, err := cat.Explain(name)
	if err != nil {
		return fmt.Errorf("%w: %q; use list_checks to see valid check names", errCheckUnknown, name)
	}
	if detail.Experimental {
		return fmt.Errorf(
			"%w: %q is only included when a scan runs with SCORECARD_EXPERIMENTAL set",
			errCheckExperimental, name)
	}
	for _, oc := range p.Capabilities().OmittedChecks {
		if strings.EqualFold(oc, name) {
			return fmt.Errorf(
				"%w: %s omits %q from its results; live scanning will cover it in a future version",
				errCheckOmitted, p.Capabilities().Source, name)
		}
	}
	return fmt.Errorf("%w: no result for %q on %s with this provider", errCheckOmitted, name, repo)
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

func toRepoScore(res *model.Result) repoScoreOutput {
	summaries := make([]checkSummary, 0, len(res.Checks))
	for _, c := range res.Checks {
		summaries = append(summaries, checkSummary{Name: c.Name, Score: c.Score, Reason: c.Reason})
	}
	return repoScoreOutput{
		Repo:        res.Repo,
		Commit:      res.Commit,
		Date:        res.Date,
		Scorecard:   res.Scorecard,
		Source:      res.Source,
		Score:       res.Score,
		Checks:      summaries,
		Caveats:     res.Caveats,
		Attribution: res.Attribution,
		Complete:    res.Complete,
	}
}
