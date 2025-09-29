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

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	pmc "github.com/ossf/scorecard/v5/cmd/internal/packagemanager"
	docs "github.com/ossf/scorecard/v5/docs/checks"
	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
	"github.com/ossf/scorecard/v5/policy"
)

type server struct {
	logger *log.Logger
}

type scorecardRequest struct {
	Repo            string   `json:"repo"`
	NPM             string   `json:"npm,omitempty"`
	PyPI            string   `json:"pypi,omitempty"`
	RubyGems        string   `json:"rubygems,omitempty"`
	Nuget           string   `json:"nuget,omitempty"`
	Commit          string   `json:"commit,omitempty"`
	FileMode        string   `json:"file_mode,omitempty"`
	Checks          []string `json:"checks,omitempty"`
	Probes          []string `json:"probes,omitempty"`
	CommitDepth     int      `json:"commit_depth,omitempty"`
	ShowDetails     bool     `json:"show_details,omitempty"`
	ShowAnnotations bool     `json:"show_annotations,omitempty"`
}

func newServer(logger *log.Logger) *server {
	return &server{
		logger: logger,
	}
}

func (s *server) handleScorecard(w http.ResponseWriter, r *http.Request) {
	var req scorecardRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		req.Repo = r.URL.Query().Get("repo")
		req.NPM = r.URL.Query().Get("npm")
		req.PyPI = r.URL.Query().Get("pypi")
		req.RubyGems = r.URL.Query().Get("rubygems")
		req.Nuget = r.URL.Query().Get("nuget")
		req.Checks = strings.Split(r.URL.Query().Get("checks"), ",")
		req.Commit = r.URL.Query().Get("commit")
		req.ShowDetails = r.URL.Query().Get("show_details") == "true"
		req.ShowAnnotations = r.URL.Query().Get("show_annotations") == "true"
		req.Probes = strings.Split(r.URL.Query().Get("probes"), ",")
		req.FileMode = r.URL.Query().Get("file_mode")
	}

	// Create a new options instance for each request to avoid race conditions
	opts := options.New()

	// Set options from request
	opts.Repo = req.Repo
	opts.NPM = req.NPM
	opts.PyPI = req.PyPI
	opts.RubyGems = req.RubyGems
	opts.Nuget = req.Nuget
	opts.Commit = req.Commit
	if opts.Commit == "" {
		opts.Commit = clients.HeadSHA
	}
	opts.CommitDepth = req.CommitDepth
	opts.ShowDetails = req.ShowDetails
	opts.ShowAnnotations = req.ShowAnnotations
	opts.LogLevel = "info"
	opts.FileMode = req.FileMode
	if opts.FileMode == "" {
		opts.FileMode = options.FileModeArchive
	}
	opts.ChecksToRun = req.Checks

	if len(req.Checks) == 1 && req.Checks[0] == "" {
		opts.ChecksToRun = nil
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid options: %v", err), http.StatusBadRequest)
		return
	}

	p := &pmc.PackageManagerClient{}
	// Set repo from package managers
	pkgResp, err := fetchGitRepositoryFromPackageManagers(opts.NPM, opts.PyPI, opts.RubyGems, opts.Nuget, p)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetchGitRepositoryFromPackageManagers: %v", err), http.StatusInternalServerError)
		return
	}
	if pkgResp.exists {
		opts.Repo = pkgResp.associatedRepo
	}

	ctx := r.Context()

	var repo clients.Repo

	repo, err = makeRepo(opts.Repo)
	if err != nil {
		http.Error(w, fmt.Sprintf("making remote repo: %v", err), http.StatusInternalServerError)
		return
	}

	// Read docs
	checkDocs, err := docs.Read()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read yaml file: %v", err), http.StatusInternalServerError)
		return
	}

	var requiredRequestTypes []checker.RequestType
	if opts.Local != "" {
		requiredRequestTypes = append(requiredRequestTypes, checker.FileBased)
	}
	if !strings.EqualFold(opts.Commit, clients.HeadSHA) {
		requiredRequestTypes = append(requiredRequestTypes, checker.CommitBased)
	}
	enabledChecks, err := policy.GetEnabled(nil, opts.Checks(), requiredRequestTypes)
	if err != nil {
		http.Error(w, fmt.Sprintf("GetEnabled: %v", err), http.StatusInternalServerError)
		return
	}

	checks := make([]string, 0, len(enabledChecks))
	for c := range enabledChecks {
		checks = append(checks, c)
	}

	enabledProbes := opts.Probes()

	scorecardOpts := []scorecard.Option{
		scorecard.WithLogLevel(log.ParseLevel(opts.LogLevel)),
		scorecard.WithCommitSHA(opts.Commit),
		scorecard.WithCommitDepth(opts.CommitDepth),
		scorecard.WithProbes(enabledProbes),
		scorecard.WithChecks(checks),
	}

	if strings.EqualFold(opts.FileMode, options.FileModeGit) {
		scorecardOpts = append(scorecardOpts, scorecard.WithFileModeGit())
	}

	repoResult, err := scorecard.Run(ctx, repo, scorecardOpts...)
	if err != nil {
		s.logger.Error(err, "scorecard.Run")
		http.Error(w, fmt.Sprintf("scorecard.Run: %v", err), http.StatusInternalServerError)
		return
	}

	repoResult.Metadata = append(repoResult.Metadata, opts.Metadata...)

	// Sort by name
	sort.Slice(repoResult.Checks, func(i, j int) bool {
		return repoResult.Checks[i].Name < repoResult.Checks[j].Name
	})

	// Return results
	w.Header().Set("Content-Type", "application/json")
	if err := repoResult.AsJSON2(w, checkDocs, &scorecard.AsJSON2ResultOption{
		LogLevel:    log.ParseLevel(opts.LogLevel),
		Details:     opts.ShowDetails,
		Annotations: opts.ShowAnnotations,
	}); err != nil {
		s.logger.Error(err, "writing JSON response")
		http.Error(w, fmt.Sprintf("failed to format results: %v", err), http.StatusInternalServerError)
		return
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stdlog.Printf("panic: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		stdlog.Printf("request completed method=%s path=%s duration=%v", r.Method, r.URL.Path, time.Since(start))
	})
}

func serveCmd(o *options.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve the scorecard program over http",
		Long:  `Start an HTTP server to run scorecard checks on repositories with REST API support.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.NewLogger(log.ParseLevel(o.LogLevel))
			srv := newServer(logger)

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet || r.Method == http.MethodPost {
					srv.handleScorecard(w, r)
				} else {
					http.NotFound(w, r)
				}
			})
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := loggerMiddleware(recoverMiddleware(corsMiddleware(mux)))

			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			httpServer := &http.Server{
				Addr:              fmt.Sprintf("0.0.0.0:%s", port),
				Handler:           handler,
				ReadHeaderTimeout: 10 * time.Second,
			}

			done := make(chan os.Signal, 1)
			signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				logger.Info("Server starting on port " + port)
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error(err, "server error")
				}
			}()

			<-done
			logger.Info("Shutting down server...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(ctx); err != nil {
				return fmt.Errorf("server shutdown: %w", err)
			}

			return nil
		},
	}
}
