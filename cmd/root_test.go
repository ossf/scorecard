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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks"
	"github.com/ossf/scorecard/v5/clients"
	docs "github.com/ossf/scorecard/v5/docs/checks"
	"github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

func TestMakeRepoLegacyAzureDevOpsURL(t *testing.T) {
	t.Setenv("SCORECARD_EXPERIMENTAL", "1")

	repo, err := makeRepo(
		"https://dnceng-public.visualstudio.com/DefaultCollection/public/_git/public",
	)
	if err != nil {
		t.Fatalf("makeRepo() error = %v", err)
	}
	if got, want := repo.Type(), clients.RepoTypeAzureDevOps; got != want {
		t.Errorf("Type() = %q, want %q", got, want)
	}
	if got, want := repo.URI(), "dev.azure.com/dnceng-public/public/_git/public"; got != want {
		t.Errorf("URI() = %q, want %q", got, want)
	}
}

func TestMakeRepoRejectsMalformedAzureDevOpsURL(t *testing.T) {
	t.Setenv("SCORECARD_EXPERIMENTAL", "1")

	tests := []string{
		"http://teamgitlab.visualstudio.com/project/_git/repo",
		"https://teamgitlab.visualstudio.com:8443/project/_git/repo",
	}
	for _, repoURI := range tests {
		if repo, err := makeRepo(repoURI); err == nil || repo != nil {
			t.Errorf("makeRepo(%q) = %v, %v; want nil repo and error", repoURI, repo, err)
		}
	}
}

//nolint:paralleltest // Captures process-wide standard error.
func TestRootCmdFormattingFailure(t *testing.T) {
	opts := options.New()
	opts.Local = t.TempDir()
	opts.ChecksToRun = []string{checks.CheckBinaryArtifacts}
	opts.Format = options.FormatJSON
	opts.ResultsFile = t.TempDir()

	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	originalStderr := os.Stderr
	os.Stderr = stderr
	t.Cleanup(func() {
		os.Stderr = originalStderr
		if err := stderr.Close(); err != nil {
			t.Error(err)
		}
	})

	err = rootCmd(opts)
	os.Stderr = originalStderr
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) || pathErr.Path != opts.ResultsFile {
		t.Errorf("rootCmd() error = %v, want output path error for %q", err, opts.ResultsFile)
	}
	diagnostics, readErr := os.ReadFile(stderr.Name())
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(diagnostics), "Failed to format results for "+opts.Local+":") {
		t.Errorf("formatting diagnostic missing from %q", diagnostics)
	}
	if strings.Contains(string(diagnostics), "Skipping "+opts.Local+":") {
		t.Errorf("completed scan reported as skipped: %q", diagnostics)
	}
}

func TestProcessRepos(t *testing.T) {
	t.Parallel()
	_, parseErr := makeRepo("::bad::")
	if parseErr == nil {
		t.Fatal("invalid repository parsed successfully")
	}
	scanErr := errors.New("scan failed")
	formatErr := errors.New("format failed")
	runtimeResult := &scorecard.Result{
		Checks: []checker.CheckResult{{Error: errors.New("check failed")}},
	}
	for _, tt := range []struct {
		result       *scorecard.Result
		err          error
		secondErr    error
		name         string
		checkFailure bool
	}{
		{
			name: "low score",
			result: &scorecard.Result{
				Checks: []checker.CheckResult{{Score: checker.MinResultScore}},
			},
		},
		{
			name: "inconclusive score",
			result: &scorecard.Result{
				Checks: []checker.CheckResult{{Score: checker.InconclusiveResultScore}},
			},
		},
		{name: "unparseable repository", err: parseErr},
		{name: "scan error", err: scanErr},
		{name: "multiple failures", err: parseErr, secondErr: scanErr},
		{name: "format error with result", result: &scorecard.Result{}, err: formatErr},
		{name: "check runtime error", result: runtimeResult, checkFailure: true},
		{name: "format and check errors", result: runtimeResult, err: formatErr, checkFailure: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uris := []string{"github.com/owner/first", "github.com/owner/second", "github.com/owner/last"}
			var visited []string
			err := processRepos(uris, func(uri string) (*scorecard.Result, error) {
				visited = append(visited, uri)
				switch uri {
				case uris[0]:
					return tt.result, tt.err
				case uris[1]:
					if tt.secondErr != nil {
						return nil, tt.secondErr
					}
				}
				return &scorecard.Result{}, nil
			})
			if !slices.Equal(visited, uris) {
				t.Errorf("processed %v, want %v", visited, uris)
			}
			for i, cause := range []error{tt.err, tt.secondErr} {
				if cause != nil && (!errors.Is(err, cause) || !strings.Contains(err.Error(), uris[i])) {
					t.Errorf("error = %v, want cause %v and repository %q", err, cause, uris[i])
				}
			}
			if errors.Is(err, errChecksFailed) != tt.checkFailure {
				t.Errorf("error = %v, want check failure = %t", err, tt.checkFailure)
			}
			if tt.err == nil && tt.secondErr == nil && !tt.checkFailure && err != nil {
				t.Errorf("successful execution returned error: %v", err)
			}
		})
	}
}

func TestProcessReposResultsFileAppend(t *testing.T) {
	t.Parallel()
	outPath := filepath.Join(t.TempDir(), "results.json")
	opts := options.New()
	opts.ResultsFile = outPath
	opts.Format = options.FormatJSON

	uris := []string{"github.com/owner/first", "github.com/owner/second"}
	checkDocs, err := docs.Read()
	if err != nil {
		t.Fatalf("docs.Read: %v", err)
	}

	err = processRepos(uris, func(uri string) (*scorecard.Result, error) {
		res := &scorecard.Result{
			Repo: scorecard.RepoInfo{Name: uri},
		}
		if formatErr := scorecard.FormatResults(opts, res, checkDocs, nil); formatErr != nil {
			return nil, formatErr
		}
		return res, nil
	})
	if err != nil {
		t.Fatalf("processRepos() error = %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("os.ReadFile: %v", err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	var repos []string
	for {
		var js scorecard.JSONScorecardResultV2
		if err := dec.Decode(&js); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("decode json: %v", err)
		}
		repos = append(repos, js.Repo.Name)
	}
	if !slices.Equal(repos, uris) {
		t.Errorf("got repos %v, want %v", repos, uris)
	}
}
