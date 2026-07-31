// Package provider retrieves OpenSSF Scorecard results from a swappable backend.
//
// The Provider interface is the central extensibility seam: this package ships a
// CachedRESTProvider (reading the public cached REST API), and a live in-process
// provider can be added later without changing the tools that depend on Provider.
package provider

import (
	"context"
	"fmt"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/model"
)

// Provider retrieves Scorecard results for a repository. A commit of "" requests
// the latest available results.
type Provider interface {
	GetResult(ctx context.Context, ref model.RepoRef, commit string) (*model.Result, error)
	Capabilities() Capabilities
}

// Capabilities describes what a provider can supply, for caveat reporting.
type Capabilities struct {
	Source        model.Source
	Complete      bool
	OmittedChecks []string
}

// NotFoundError indicates the provider has no results for the repository.
type NotFoundError struct {
	Ref model.RepoRef
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf(
		"no published OpenSSF Scorecard results for %s. The repository may not have "+
			"opted into publishing results (publish_results: true), or it may not be "+
			"covered by the weekly public scan. On-demand live scanning will be "+
			"available in a future version.",
		e.Ref.String(),
	)
}
