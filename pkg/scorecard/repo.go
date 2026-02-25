package scorecard

import (
	"strings"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/azuredevopsrepo"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
)

type RepoType string

const (
	RepoUnknown     RepoType = "unknown"
	RepoLocal       RepoType = "local"
	RepoGitLocal    RepoType = "git-local" // is not supported by any check yet.
	RepoGitHub      RepoType = "github"
	RepoGitLab      RepoType = "gitlab"
	RepoAzureDevOPs RepoType = "azuredevops"
)

func GetRepoType(repo clients.Repo) RepoType {
	switch repo.(type) {
	case *localdir.Repo:
		return RepoLocal
	case *githubrepo.Repo:
		return RepoGitHub
	case *gitlabrepo.Repo:
		return RepoGitLab
	case *azuredevopsrepo.Repo:
		return RepoAzureDevOPs
	default:
		return RepoUnknown
	}
}

func RepoTypeFromString(repo string) RepoType {
	rt := RepoType(strings.ToLower(strings.TrimSpace(repo)))
	switch rt {
	case RepoLocal, RepoGitLocal, RepoGitHub, RepoGitLab, RepoAzureDevOPs:
		return rt
	default:
		return RepoUnknown
	}
}
