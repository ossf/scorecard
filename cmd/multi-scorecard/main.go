// Copyright (c) Jeff Mendoza <jlm@jlm.name>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v63/github"

	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/docs/checks"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

func main() {
	var appID = flag.Int64("appid", 0, "")
	var keyFile = flag.String("keyfile", "", "")
	flag.Parse()

	if err := do(*appID, *keyFile); err != nil {
		log.Fatal(err)
	}
}

func do(appID int64, keyFile string) error {
	dt := http.DefaultTransport
	ctx := context.Background()

	// Read in GitHub App private key
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return err
	}

	// Create authenticated transport for App
	at, err := ghinstallation.NewAppsTransport(dt, appID, key)
	if err != nil {
		return err
	}

	// Query GitHub for list of installations for this App
	ghc := github.NewClient(&http.Client{Transport: at})
	insts, _, err := ghc.Apps.ListInstallations(ctx, &github.ListOptions{PerPage: 100})
	if err != nil {
		return err
	}

	checkDocs, err := checks.Read()
	if err != nil {
		return fmt.Errorf("cannot read yaml file: %w", err)
	}

	// Iterate through installations
	var results [][]byte
	for _, inst := range insts {
		// Create a new authenticated transport for this App installation that will
		// be used for Scorecard.
		it := ghinstallation.NewFromAppsTransport(at, inst.GetID())

		// Query GitHub for list of repos available to this installation
		ic := github.NewClient(&http.Client{Transport: it})
		repos, _, err := ic.Apps.ListRepos(ctx, &github.ListOptions{PerPage: 100})
		if err != nil {
			return err
		}

		// Create Scorecard RepoClient using authenticated transport
		rc := githubrepo.CreateGithubRepoClientWithTransport(ctx, it)

		// Iterate through installations
		for _, repo := range repos.Repositories {
			// Create Scorecard Repo object for repo name we want to scan
			screpo, err := githubrepo.MakeGithubRepo(repo.GetFullName())
			if err != nil {
				return err
			}

			// Run scorecard with Repo and RepoClient
			res, err := scorecard.Run(ctx, screpo,
				scorecard.WithRepoClient(rc),
			)
			if err != nil {
				return err
			}

			// Write results as json into buffer
			var b bytes.Buffer
			if err := res.AsJSON2(&b, checkDocs, nil); err != nil {
				return err
			}
			results = append(results, b.Bytes())
		}
	}

	// Output results as a JSON array
	for i, r := range results {
		if i == 0 {
			fmt.Print("[")
		}
		fmt.Print(string(r))
		if i == len(results)-1 {
			fmt.Print("]")
		} else {
			fmt.Print(",")
		}
	}
	return nil
}
