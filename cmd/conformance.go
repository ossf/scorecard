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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/localdir"
	"github.com/ossf/scorecard/v5/internal/conformance"
	"github.com/ossf/scorecard/v5/internal/conformance/catalog"
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

const experimentalWarning = "WARNING: conformance mode (--catalog) is experimental. " +
	"Its options and output may change or be removed in a future release."

var errNoMappedProbes = errors.New("no mapped requirement applies to the profile, so there is nothing to evaluate")

// runConformance evaluates the selected catalog release for one repository
// and writes the results. It runs only the probes the release's mappings
// name, and computes no check scores.
func runConformance(ctx context.Context, o *options.Options, uri string) error {
	fmt.Fprintln(os.Stderr, experimentalWarning)

	release, err := catalog.Lookup(o.Catalog)
	if err != nil {
		return fmt.Errorf("catalog: %w", err)
	}
	probes := release.Probes(o.Profile)
	if len(probes) == 0 {
		// scorecard.Run with no probes would run every check instead.
		return fmt.Errorf("%s %s, %s: %w", release.Catalog, release.Version, o.Profile, errNoMappedProbes)
	}

	var repo clients.Repo
	if o.Local != "" {
		repo, err = localdir.MakeLocalDirRepo(uri)
		if err != nil {
			return fmt.Errorf("localdir: %w", err)
		}
	} else {
		repo, err = makeRepo(uri)
		if err != nil {
			return err
		}
	}

	opts := []scorecard.Option{
		scorecard.WithLogLevel(sclog.ParseLevel(o.LogLevel)),
		scorecard.WithCommitSHA(o.Commit),
		scorecard.WithCommitDepth(o.CommitDepth),
		scorecard.WithProbes(probes),
	}
	if strings.EqualFold(o.FileMode, options.FileModeGit) {
		opts = append(opts, scorecard.WithFileModeGit())
	}
	result, err := scorecard.Run(ctx, repo, opts...)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	report := conformanceReport(&result, release, o.Profile)

	output := os.Stdout
	if o.ResultsFile != "" {
		output, err = os.Create(o.ResultsFile)
		if err != nil {
			return fmt.Errorf("unable to create output file: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Writing results to", o.ResultsFile)
		defer output.Close()
	}
	return writeConformance(output, o.Format, report)
}

// conformanceReport evaluates the release's mappings for the profile from
// the scan's probe findings.
func conformanceReport(result *scorecard.Result, release *conformance.Release, profile string) *conformance.Report {
	in := conformance.Input{
		// Scorecard records the default branch for every scan; it is
		// "unknown" when the platform can't report it, which matches no
		// finding and so yields UNKNOWN.
		DefaultBranch: result.RawResults.Metadata.Metadata["repository.defaultBranch"],
		Findings:      result.Findings,
	}
	results, notEvaluated := conformance.Evaluate(release, profile, in)
	return &conformance.Report{
		Date:             result.Date,
		Release:          release,
		Repo:             result.Repo.Name,
		Commit:           result.Repo.CommitSHA,
		ScorecardVersion: result.Scorecard.Version,
		ScorecardCommit:  result.Scorecard.CommitSHA,
		Profile:          profile,
		Results:          results,
		NotEvaluated:     notEvaluated,
	}
}

func writeConformance(w io.Writer, format string, report *conformance.Report) error {
	var err error
	switch format {
	case options.FormatJSON:
		err = report.AsJSON(w)
	case options.FormatGemara:
		err = report.AsGemara(w)
	default:
		// Option validation rejects other formats in conformance mode.
		return fmt.Errorf("format %q: %w", format, errCatalogFormat)
	}
	if err != nil {
		return fmt.Errorf("format results: %w", err)
	}
	return nil
}

var errCatalogFormat = errors.New("conformance mode writes only json or gemara")
