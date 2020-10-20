package cmd

import (
	"net/http"

	"github.com/dlorenc/scorecard/checks"
	"github.com/dlorenc/scorecard/pkg"
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
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

		http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			// r := r.URL.Query().Get("repo")
			var repo pkg.RepoURL
			ctx := r.Context()
			pkg.RunScorecards(ctx, sugar, repo, checks.AllChecks)

		})
		http.ListenAndServe("localhost:8080", nil)
	},
}
