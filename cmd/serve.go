// Copyright 2020 OpenSSF Scorecard Authors
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
	"github.com/ossf/scorecard/v5/clients/localdir"
	pmc "github.com/ossf/scorecard/v5/cmd/internal/packagemanager"
	docs "github.com/ossf/scorecard/v5/docs/checks"
	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
	"github.com/ossf/scorecard/v5/policy"
)

type server struct {
	logger *log.Logger
	opts   *options.Options
}

type scorecardRequest struct {
	Repo        string   `json:"repo"`
	Local       string   `json:"local,omitempty"`
	NPM         string   `json:"npm,omitempty"`
	PyPI        string   `json:"pypi,omitempty"`
	RubyGems    string   `json:"rubygems,omitempty"`
	Nuget       string   `json:"nuget,omitempty"`
	Checks      []string `json:"checks,omitempty"`
	Commit      string   `json:"commit,omitempty"`
	CommitDepth int      `json:"commit_depth,omitempty"`
	ShowDetails bool     `json:"show_details,omitempty"`
	Format      string   `json:"format,omitempty"`
	LogLevel    string   `json:"log_level,omitempty"`
	Probes      []string `json:"probes,omitempty"`
	FileMode    string   `json:"file_mode,omitempty"`
	PolicyFile  string   `json:"policy_file,omitempty"`
}

func newServer(logger *log.Logger, opts *options.Options) *server {
	return &server{
		logger: logger,
		opts:   opts,
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
		req.Local = r.URL.Query().Get("local")
		req.NPM = r.URL.Query().Get("npm")
		req.PyPI = r.URL.Query().Get("pypi")
		req.RubyGems = r.URL.Query().Get("rubygems")
		req.Nuget = r.URL.Query().Get("nuget")
		req.Checks = strings.Split(r.URL.Query().Get("checks"), ",")
		req.Commit = r.URL.Query().Get("commit")
		req.ShowDetails = r.URL.Query().Get("show_details") == "true"
		req.Format = r.URL.Query().Get("format")
		req.LogLevel = r.URL.Query().Get("log_level")
		req.Probes = strings.Split(r.URL.Query().Get("probes"), ",")
		req.FileMode = r.URL.Query().Get("file_mode")
		req.PolicyFile = r.URL.Query().Get("policy_file")
	}

	// Set options
	s.opts.Repo = req.Repo
	s.opts.Local = req.Local
	s.opts.NPM = req.NPM
	s.opts.PyPI = req.PyPI
	s.opts.RubyGems = req.RubyGems
	s.opts.Nuget = req.Nuget
	s.opts.Commit = req.Commit
	if s.opts.Commit == "" {
		s.opts.Commit = clients.HeadSHA
	}
	s.opts.CommitDepth = req.CommitDepth
	s.opts.ShowDetails = req.ShowDetails
	s.opts.Format = req.Format
	if s.opts.Format == "" {
		s.opts.Format = options.FormatDefault
	}
	s.opts.LogLevel = req.LogLevel
	if s.opts.LogLevel == "" {
		s.opts.LogLevel = "info"
	}
	s.opts.FileMode = req.FileMode
	if s.opts.FileMode == "" {
		s.opts.FileMode = options.FileModeArchive
	} else if s.opts.FileMode != options.FileModeArchive && s.opts.FileMode != options.FileModeGit {
		http.Error(w, fmt.Sprintf("unsupported file mode: %s", s.opts.FileMode), http.StatusBadRequest)
		return
	}
	s.opts.PolicyFile = req.PolicyFile
	s.opts.ChecksToRun = req.Checks

	// Validate options
	if err := s.opts.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid options: %v", err), http.StatusBadRequest)
		return
	}

	p := &pmc.PackageManagerClient{}
	// Set repo from package managers
	pkgResp, err := fetchGitRepositoryFromPackageManagers(s.opts.NPM, s.opts.PyPI, s.opts.RubyGems, s.opts.Nuget, p)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetchGitRepositoryFromPackageManagers: %v", err), http.StatusInternalServerError)
		return
	}
	if pkgResp.exists {
		s.opts.Repo = pkgResp.associatedRepo
	}

	pol, err := policy.ParseFromFile(s.opts.PolicyFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("readPolicy: %v", err), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	var repo clients.Repo
	if s.opts.Local != "" {
		repo, err = localdir.MakeLocalDirRepo(s.opts.Local)
		if err != nil {
			http.Error(w, fmt.Sprintf("making local dir: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		repo, err = makeRepo(s.opts.Repo)
		if err != nil {
			http.Error(w, fmt.Sprintf("making remote repo: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Read docs
	checkDocs, err := docs.Read()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read yaml file: %v", err), http.StatusInternalServerError)
		return
	}

	var requiredRequestTypes []checker.RequestType
	if s.opts.Local != "" {
		requiredRequestTypes = append(requiredRequestTypes, checker.FileBased)
	}
	if !strings.EqualFold(s.opts.Commit, clients.HeadSHA) {
		requiredRequestTypes = append(requiredRequestTypes, checker.CommitBased)
	}

	enabledChecks, err := policy.GetEnabled(pol, s.opts.Checks(), requiredRequestTypes)
	stdlog.Printf("DEBUG: enabledChecks = %#v", enabledChecks)
	if err != nil {
		http.Error(w, fmt.Sprintf("GetEnabled: %v", err), http.StatusInternalServerError)
		return
	}

	checks := make([]string, 0, len(enabledChecks))
	for c := range enabledChecks {
		checks = append(checks, c)
	}

	enabledProbes := s.opts.Probes()

	opts := []scorecard.Option{
		scorecard.WithLogLevel(log.ParseLevel(s.opts.LogLevel)),
		scorecard.WithCommitSHA(s.opts.Commit),
		scorecard.WithCommitDepth(s.opts.CommitDepth),
		scorecard.WithProbes(enabledProbes),
		scorecard.WithChecks(checks),
	}

	if strings.EqualFold(s.opts.FileMode, options.FileModeGit) {
		opts = append(opts, scorecard.WithFileModeGit())
	}

	repoResult, err := scorecard.Run(ctx, repo, opts...)
	if err != nil {
		s.logger.Error(err, "scorecard.Run")
		http.Error(w, fmt.Sprintf("scorecard.Run: %v", err), http.StatusInternalServerError)
		return
	}

	repoResult.Metadata = append(repoResult.Metadata, s.opts.Metadata...)

	// Sort by name
	sort.Slice(repoResult.Checks, func(i, j int) bool {
		return repoResult.Checks[i].Name < repoResult.Checks[j].Name
	})

	// Return results
	w.Header().Set("Content-Type", "application/json")
	if err := repoResult.AsJSON2(w, checkDocs, &scorecard.AsJSON2ResultOption{
		LogLevel:    log.ParseLevel(s.opts.LogLevel),
		Details:     s.opts.ShowDetails,
		Annotations: false,
	}); err != nil {
		s.logger.Error(err, "writing JSON response")
		http.Error(w, fmt.Sprintf("failed to format results: %v", err), http.StatusInternalServerError)
		return
	}
}

// CORS middleware for net/http
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

// Recover middleware for net/http
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

// Logger middleware for net/http
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		stdlog.Printf("%s %s %s", r.Method, r.URL, time.Since(start))
	})
}

func serveCmd(o *options.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve the scorecard program over http",
		Long:  `Start an HTTP server to run scorecard checks on repositories with REST API support.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.NewLogger(log.ParseLevel(o.LogLevel))
			srv := newServer(logger, o)

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

			// Compose middlewares: logger -> recover -> cors -> mux
			handler := loggerMiddleware(recoverMiddleware(corsMiddleware(mux)))

			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			httpServer := &http.Server{
				Addr:    fmt.Sprintf("0.0.0.0:%s", port),
				Handler: handler,
			}

			// Graceful shutdown
			done := make(chan os.Signal, 1)
			signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				logger.Info("Server starting on port " + port)
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
