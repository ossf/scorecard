package runner

import (
	"context"
	"fmt"
	"os"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/ossfuzz"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

const (
	commit      = clients.HeadSHA
	commitDepth = 0 // default
)

type Runner struct {
	ctx           context.Context
	enabledChecks checker.CheckNameToFnMap
	repoClient    clients.RepoClient
	ossFuzz       clients.RepoClient
	cii           clients.CIIBestPracticesClient
	vuln          clients.VulnerabilitiesClient
}

func New() Runner {
	ctx := context.Background()
	logger := log.NewLogger(log.DefaultLevel)
	return Runner{
		ctx:           ctx,
		repoClient:    githubrepo.CreateGithubRepoClient(ctx, logger),
		ossFuzz:       ossfuzz.CreateOSSFuzzClient(ossfuzz.StatusURL),
		cii:           clients.DefaultCIIBestPracticesClient(),
		vuln:          clients.DefaultVulnerabilitiesClient(),
		enabledChecks: checks.GetAll(),
	}
}

//nolint:wrapcheck
func (r *Runner) Run(repoURI string) (pkg.ScorecardResult, error) {
	fmt.Fprintf(os.Stdout, "running for repo: %v\n", repoURI)
	// TODO (gitlab?)
	repo, err := githubrepo.MakeGithubRepo(repoURI)
	if err != nil {
		return pkg.ScorecardResult{}, err
	}
	return pkg.RunScorecard(r.ctx, repo, commit, commitDepth, r.enabledChecks, r.repoClient, r.ossFuzz, r.cii, r.vuln)
}
