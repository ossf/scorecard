package depdiff

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	gogh "github.com/google/go-github/v38/github"
)

type DepDiffContext struct {
	OwnerName   string
	RepoName    string
	BaseSHA     string
	HeadSHA     string
	AccessToken string
}

type DepDiffFormat string

const (
	JSON      DepDiffFormat = "json"
	Plaintext DepDiffFormat = "plaintext"
	Markdown  DepDiffFormat = "md"
)

func (f *DepDiffFormat) IsValid() bool {
	switch *f {
	case JSON, Plaintext, Markdown:
		return true
	default:
		return false
	}
}

func GetDependencyDiff(
	ownerName, repoName, baseSHA, headSHA, accessToken string,
	format DepDiffFormat,
) (interface{}, error) {
	format = DepDiffFormat(strings.ToLower(string(format)))
	if !format.IsValid() {
		return nil, ErrInvalidDepDiffFormat
	}
	ctx := DepDiffContext{
		OwnerName:   ownerName,
		RepoName:    repoName,
		BaseSHA:     baseSHA,
		HeadSHA:     headSHA,
		AccessToken: accessToken,
	}

	// Fetch dependency diffs using the GitHub Dependency Review API.
	deps, err := FetchDepDiffDataFromGitHub(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(deps)

	switch format {
	case JSON:
		return nil, nil
	case Markdown:
		return nil, ErrDepDiffFormatNotSupported
	case Plaintext:
		return nil, ErrDepDiffFormatNotSupported
	default:
		return nil, ErrDepDiffFormatNotSupported
	}
}

// Get the depednency-diff using the GitHub Dependency Review
// (https://docs.github.com/en/rest/dependency-graph/dependency-review) API
func FetchDepDiffDataFromGitHub(ctx DepDiffContext) ([]Dependency, error) {
	// Set a ten-seconds timeout to make sure the client can be created correctly.
	client := gogh.NewClient(&http.Client{Timeout: 10 * time.Second})
	reqURL := path.Join(
		"repos", ctx.OwnerName, ctx.RepoName, "dependency-graph", "compare",
		ctx.BaseSHA+"..."+ctx.HeadSHA,
	)
	req, err := client.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	// To specify the return type to be JSON.
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	// An access token is required in the request header to be able to use this API.
	req.Header.Set("Authorization", "token "+ctx.AccessToken)

	depDiff := []Dependency{}
	_, err = client.Do(req.Context(), req, &depDiff)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	return depDiff, nil
}

func GetAggregateScore(d Dependency) (float32, error) {
	return 0, nil
}
