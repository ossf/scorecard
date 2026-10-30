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
	"io"
	stdlog "log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
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

const (
	defaultListenHost = "127.0.0.1"
	defaultPort       = "8080"
	// maxRequestBodyBytes caps POST bodies; a scan request is well under 1 KiB.
	maxRequestBodyBytes = 1 << 20
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

func parsePOSTRequest(w http.ResponseWriter, r *http.Request) (*scorecardRequest, bool) {
	var req scorecardRequest
	body := http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	err := json.NewDecoder(body).Decode(&req)
	if err == nil {
		// The decoder stops after the first JSON value; read the rest so an
		// oversized body is rejected even when it starts with valid JSON.
		_, err = io.Copy(io.Discard, body)
	}
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return nil, false
	}
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return nil, false
	}
	return &req, true
}

func parseGETRequest(w http.ResponseWriter, r *http.Request) (*scorecardRequest, bool) {
	q := r.URL.Query()
	req := scorecardRequest{
		Repo:            q.Get("repo"),
		NPM:             q.Get("npm"),
		PyPI:            q.Get("pypi"),
		RubyGems:        q.Get("rubygems"),
		Nuget:           q.Get("nuget"),
		Checks:          strings.Split(q.Get("checks"), ","),
		Commit:          q.Get("commit"),
		ShowDetails:     q.Get("show_details") == "true",
		ShowAnnotations: q.Get("show_annotations") == "true",
		Probes:          strings.Split(q.Get("probes"), ","),
		FileMode:        q.Get("file_mode"),
	}
	if depthStr := q.Get("commit_depth"); depthStr != "" {
		depth, err := strconv.Atoi(depthStr)
		if err != nil {
			http.Error(w, "invalid commit_depth", http.StatusBadRequest)
			return nil, false
		}
		req.CommitDepth = depth
	}
	return &req, true
}

func (s *server) handleScorecard(w http.ResponseWriter, r *http.Request) {
	var (
		req *scorecardRequest
		ok  bool
	)
	if r.Method == http.MethodPost {
		req, ok = parsePOSTRequest(w, r)
	} else {
		req, ok = parseGETRequest(w, r)
	}
	if !ok {
		return
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
		http.Error(w, fmt.Sprintf("making remote repo: %v", err), http.StatusBadRequest)
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
	enabledChecks, err := policy.GetEnabled(nil, opts.Checks(), requiredRequestTypes, repo.Type())
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

// isCrossOrigin applies the rule of http.CrossOriginProtection to every
// method. The standard library exempts GET, but a GET here starts a scan.
func isCrossOrigin(r *http.Request) bool {
	switch r.Header.Get("Sec-Fetch-Site") {
	case "same-origin", "none":
		return false
	case "":
		// Older browsers send only Origin; non-browser clients send neither.
	default:
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	o, err := url.Parse(origin)
	return err != nil || o.Host != r.Host
}

func crossOriginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isCrossOrigin(r) {
			http.Error(w, "cross-origin request rejected", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// isLoopback reports whether host is localhost or a loopback IP address.
func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// loopbackHostMiddleware rejects requests that name a non-loopback host, so
// a name that resolves to a loopback address can't reach the server.
func loopbackHostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if !isLoopback(strings.Trim(host, "[]")) {
			http.Error(w, "host not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// listenAddress returns the host and the address to listen on. listen is a
// host or a host:port; the port defaults to port, then to 8080.
func listenAddress(listen, port string) (host, addr string) {
	if port == "" {
		port = defaultPort
	}
	host = defaultListenHost
	if listen != "" {
		var p string
		var err error
		host, p, err = net.SplitHostPort(listen)
		if err != nil {
			host = strings.Trim(listen, "[]")
		} else if p != "" {
			port = p
		}
	}
	return host, net.JoinHostPort(host, port)
}

func newHandler(srv *server, loopback bool) http.Handler {
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

	handler := crossOriginMiddleware(mux)
	if loopback {
		handler = loopbackHostMiddleware(handler)
	}
	return loggerMiddleware(recoverMiddleware(handler))
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
	var listen string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the scorecard program over http",
		Long: `Start an HTTP server to run scorecard checks on repositories with REST API support.

The server has no authentication or rate limiting and runs every scan with
the server's own credentials. It listens on 127.0.0.1 unless --listen names
another address. Use --listen only on a trusted network, with a token
limited to the repositories you intend to scan.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Flags are valid by now; don't print usage for a runtime error.
			cmd.SilenceUsage = true
			logger := log.NewLogger(log.ParseLevel(o.LogLevel))
			srv := newServer(logger)

			host, addr := listenAddress(listen, os.Getenv("PORT"))
			loopback := isLoopback(host)

			httpServer := &http.Server{
				Handler:           newHandler(srv, loopback),
				ReadHeaderTimeout: 10 * time.Second,
			}

			var lc net.ListenConfig
			ln, err := lc.Listen(cmd.Context(), "tcp", addr)
			if err != nil {
				return fmt.Errorf("listening on %s: %w", addr, err)
			}

			done := make(chan os.Signal, 1)
			signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

			logger.Info("Server listening on " + addr)
			if !loopback {
				logger.Info("Server has no authentication and is reachable from other hosts")
			}
			go func() {
				if err := httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	cmd.Flags().StringVar(&listen, "listen", "",
		"host or host:port to listen on (default 127.0.0.1; port from PORT, then 8080)")
	return cmd
}
