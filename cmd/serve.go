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
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
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
		logger := log.NewLogger(log.ParseLevel(flagLogLevel))

		t, err := template.New("webpage").Parse(tpl)
		if err != nil {
			// TODO(log): Should this actually panic?
			logger.Error(err, "parsing webpage template")
			panic(err)
		}

		http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			repoParam := r.URL.Query().Get("repo")
			const length = 3
			s := strings.SplitN(repoParam, "/", length)
			if len(s) != length {
				rw.WriteHeader(http.StatusBadRequest)
			}
			repo, err := githubrepo.MakeGithubRepo(repoParam)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
			}
			ctx := r.Context()
			repoClient := githubrepo.CreateGithubRepoClient(ctx, logger)
			ossFuzzRepoClient, err := githubrepo.CreateOssFuzzRepoClient(ctx, logger)
			vulnsClient := clients.DefaultVulnerabilitiesClient()
			if err != nil {
				logger.Error(err, "initializing clients")
				rw.WriteHeader(http.StatusInternalServerError)
			}
			defer ossFuzzRepoClient.Close()
			ciiClient := clients.DefaultCIIBestPracticesClient()
			repoResult, err := pkg.RunScorecards(
				ctx, repo, clients.HeadSHA /*commitSHA*/, false /*raw*/, checks.AllChecks, repoClient,
				ossFuzzRepoClient, ciiClient, vulnsClient)
			if err != nil {
				logger.Error(err, "running enabled scorecard checks on repo")
				rw.WriteHeader(http.StatusInternalServerError)
			}

			if r.Header.Get("Content-Type") == "application/json" {
				if err := repoResult.AsJSON(flagShowDetails, log.ParseLevel(flagLogLevel), rw); err != nil {
					// TODO(log): Improve error message
					logger.Error(err, "")
					rw.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
			if err := t.Execute(rw, repoResult); err != nil {
				// TODO(log): Improve error message
				logger.Error(err, "")
			}
		})
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		fmt.Printf("Listening on localhost:%s\n", port)
		err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
		if err != nil {
			// TODO(log): Should this actually panic?
			logger.Error(err, "listening and serving")
			panic(err)
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
