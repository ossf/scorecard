// Copyright 2020 Security Scorecard Authors
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
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ossf/scorecard/v2/checks"
	"github.com/ossf/scorecard/v2/clients/githubrepo"
	"github.com/ossf/scorecard/v2/pkg"
	"github.com/ossf/scorecard/v2/repos"
	"github.com/ossf/scorecard/v2/roundtripper"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the scorecard program over http",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(*logLevel)
		logger, err := cfg.Build()
		if err != nil {
			log.Fatalf("unable to construct logger: %v", err)
		}
		//nolint
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()
		t, err := template.New("webpage").Parse(tpl)
		if err != nil {
			sugar.Panic(err)
		}

		http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			repoParam := r.URL.Query().Get("repo")
			const length = 3
			s := strings.SplitN(repoParam, "/", length)
			if len(s) != length {
				rw.WriteHeader(http.StatusBadRequest)
			}
			repo := repos.RepoURL{}
			if err := repo.Set(repoParam); err != nil {
				rw.WriteHeader(http.StatusBadRequest)
			}
			sugar.Info(repoParam)
			ctx := r.Context()
			rt := roundtripper.NewTransport(ctx, sugar)
			httpClient := &http.Client{
				Transport: rt,
			}
			githubClient := github.NewClient(httpClient)
			graphClient := githubv4.NewClient(httpClient)
			repoClient := githubrepo.CreateGithubRepoClient(ctx, githubClient, graphClient)
			repoResult, err := pkg.RunScorecards(ctx, repo, checks.AllChecks, repoClient, httpClient, githubClient, graphClient)
			if err != nil {
				sugar.Error(err)
				rw.WriteHeader(http.StatusInternalServerError)
			}

			if r.Header.Get("Content-Type") == "application/json" {
				if err := repoResult.AsJSON(showDetails, *logLevel, rw); err != nil {
					sugar.Error(err)
					rw.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
			if err := t.Execute(rw, repoResult); err != nil {
				sugar.Warn(err)
			}
		})
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		fmt.Printf("Listening on localhost:%s\n", port)
		err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
		if err != nil {
			log.Fatal("ListenAndServe ", err)
		}
	},
}

const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Scorecard Results for: {{.Repo}}</title>
	</head>
	<body>
		{{range .Checks}}
			<div>
				<p>{{ .Name }}: {{ .Pass }}</p>
			</div>
		{{end}}
	</body>
</html>`
