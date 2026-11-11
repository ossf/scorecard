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
}

// NotFoundError indicates the provider has no results for the repository.
type NotFoundError struct {
	Ref model.RepoRef
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf(
		"no published OpenSSF Scorecard results for %s. The repository may not have "+
			"opted into publishing results (publish_results: true), or it may not be "+
			"covered by the weekly public scan. `scorecard --repo=%s` computes results "+
			"locally.",
		e.Ref.String(), e.Ref.String(),
	)
}
