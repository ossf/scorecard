package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/google/go-github/v42/github"
	"golang.org/x/oauth2"
)

// **************************************
// Set these parameters.
const orgName string = "organization name"
const pat string = "personal access token"

var RepoList = []string{} // Optional, leave empty to process all repos under org.
// **************************************

// Adds the OpenSSF Scorecard workflow to all repositores under the given organization.
func main() {
	// Get github user client.
	context := context.Background()
	tokenService := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tokenClient := oauth2.NewClient(context, tokenService)
	client := github.NewClient(tokenClient)

	// If not provided, get all repositories under organization.
	if len(REPO_LIST) == 0 {
		lops := &github.RepositoryListByOrgOptions{Type: "all"}
		repos, _, err := client.Repositories.ListByOrg(context, orgName, lops)
		err_check(err, "Error listing organization's repos.")

		// Convert to list of repository names.
		for _, repo := range repos {
			REPO_LIST = append(REPO_LIST, *repo.Name)
		}
	}

	// Get yml file into byte array.
	workflowContent, err := ioutil.ReadFile("scorecards-analysis.yml")
	err_check(err, "Error reading in scorecard workflow file.")

	// Process each repository.
	for _, repoName := range REPO_LIST {

		// Get repo metadata.
		repo, _, err := client.Repositories.Get(context, orgName, repoName)
		if err != nil {
			fmt.Println("Skipped repo", repoName, "because it does not exist or could not be accessed.")
			continue
		}

		// Get head commit SHA of default branch.
		defaultBranch, _, err := client.Repositories.GetBranch(context, orgName, repoName, *repo.DefaultBranch, true)

		if err != nil {
			fmt.Println("Skipped repo", repoName, "because it's default branch could not be accessed.")
			continue
		}
		defaultBranchSHA := defaultBranch.Commit.SHA

		// Skip if scorecard file already exists in workflows folder.
		scoreFileContent, _, _, err := client.Repositories.GetContents(context, orgName, repoName, ".github/workflows/scorecards-analysis.yml", &github.RepositoryContentGetOptions{})
		if scoreFileContent != nil || err == nil {
			fmt.Println("Skipped repo", repoName, "since scorecard workflow already exists.")
			continue
		}

		// Skip if branch scorecard already exists.
		scorecardBranch, _, err := client.Repositories.GetBranch(context, orgName, repoName, "scorecard", true)
		if scorecardBranch != nil || err == nil {
			fmt.Println("Skipped repo", repoName, "since branch scorecard already exists.")
			continue
		}

		// Create new branch using a reference that stores the new commit hash.
		ref := &github.Reference{
			Ref:    github.String("refs/heads/scorecard"),
			Object: &github.GitObject{SHA: defaultBranchSHA},
		}
		_, _, err = client.Git.CreateRef(context, orgName, repoName, ref)
		if err != nil {
			fmt.Println("Skipped repo", repoName, "because new branch could not be created.")
			continue
		}

		// Create file in repository.
		opts := &github.RepositoryContentFileOptions{
			Message: github.String("Adding scorecard workflow"),
			Content: []byte(workflowContent),
			Branch:  github.String("scorecard"),
		}
		_, _, err = client.Repositories.CreateFile(context, orgName, repoName, ".github/workflows/scorecards-analysis.yml", opts)
		if err != nil {
			fmt.Println("Skipped repo", repoName, "because new file could not be created.")
			continue
		}

		// Create Pull request.
		pr := &github.NewPullRequest{
			Title: github.String("Added Scorecard Workflow"),
			Head:  github.String("scorecard"),
			Base:  github.String(*defaultBranch.Name),
			Body:  github.String("Added the workflow for OpenSSF's Security Scorecard"),
			Draft: github.Bool(false),
		}

		_, _, err = client.PullRequests.Create(context, orgName, repoName, pr)
		if err != nil {
			fmt.Println("Skipped repo", repoName, "because pull request could not be created.")
			continue
		}

		// Logging.
		fmt.Println("Successfully added scorecard workflow PR from scorecard to", *defaultBranch.Name, "branch of repo", repoName)
	}
}

func err_check(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
	}
}
