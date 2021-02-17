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
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ossf/scorecard/checks"
	"github.com/ossf/scorecard/pkg"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

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
		logger, _ := cfg.Build()
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
			repo := pkg.RepoURL{}
			if err := repo.Set(repoParam); err != nil {
				rw.WriteHeader(http.StatusBadRequest)
			}
			sugar.Info(repoParam)
			ctx := r.Context()
			resultCh := pkg.RunScorecards(ctx, sugar, repo, checks.AllChecks)
			tc := tc{
				URL: repoParam,
			}
			for r := range resultCh {
				sugar.Info(r)
				tc.Results = append(tc.Results, r)
			}

			if r.Header.Get("Content-Type") == "application/json" {
				result, err := encodeJson(repo.String(), tc.Results)
				if err != nil {
					sugar.Error(err)
					rw.WriteHeader(http.StatusInternalServerError)
				}
				if _, err := rw.Write(result); err != nil {
					sugar.Error(err)
				}
				return
			}

			if err := t.Execute(rw, tc); err != nil {
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

// encodeJson encodes the result to json
func encodeJson(repo string, results []pkg.Result) ([]byte, error) {
	d := time.Now()
	or := record{
		Repo: repo,
		Date: d.Format("2006-01-02"),
	}

	for _, r := range results {
		var details []string
		if showDetails {
			details = r.Cr.Details
		}
		or.Checks = append(or.Checks, checkResult{
			CheckName:  r.Name,
			Pass:       r.Cr.Pass,
			Confidence: r.Cr.Confidence,
			Details:    details,
		})
	}
	output, err := json.Marshal(or)
	if err != nil {
		return nil, err
	}
	return output, nil
}

type tc struct {
	URL     string
	Results []pkg.Result
}

const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Scorecard Results for: {{.URL}}</title>
	</head>
	<body>
		{{range .Results}}
			<div>
				<p>{{ .Name }}: {{ .Cr.Pass }}</p>
			</div>
		{{end}}
	</body>
</html>`
