package command

import (
	"context"
	"os"

	"github.com/golang/glog"
	"github.com/ossf/scorecard-attestor/policy"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

func checkCommand() {
	ctx := context.Background()

	// Read the Binauthz attestation policy
	if policyPath == "" {
		exitOnBadFlags(SignerMode(mode), "policy path is empty")
	}
	attestationPolicy, err := policy.ParseAttestationPolicyFromFile(policyPath)
	if err != nil {
		glog.Fatalf("Fail to load scorecard attestation policy: %v", err)
	}

	if repoURL == "" {
		buildRepo := os.Getenv("REPO_NAME")
		if buildRepo == "" {
			exitOnBadFlags(SignerMode(mode), "repoURL not specified")
		}
		repoURL = buildRepo
		glog.Infof("Found repo URL %s Cloud Build environment", repoURL)
	} else {
		glog.Infof("Running scorecard on %s", repoURL)
	}

	if commitSHA == "" {
		buildSHA := os.Getenv("COMMIT_SHA")
		if buildSHA == "" {
			glog.Infof("commit not specified, running on HEAD")
		} else {
			commitSHA = buildSHA
			glog.Infof("Found revision %s Cloud Build environment", commitSHA)
		}
	}

	logger := sclog.NewLogger(sclog.DefaultLevel)

	repo, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, err := checker.GetClients(
		ctx, repoURL, "", logger)

	enabledChecks := map[string]checker.Check{
		"BinaryArtifacts": {
			Fn: checks.BinaryArtifacts,
			SupportedRequestTypes: []checker.RequestType{
				checker.CommitBased,
			},
		},
	}

	repoResult, err := pkg.RunScorecards(
		ctx,
		repo,
		commitSHA,
		enabledChecks,
		repoClient,
		ossFuzzRepoClient,
		ciiClient,
		vulnsClient,
	)
	if err != nil {
		glog.Fatalf("RunScorecards: %w", err)
	}

	result, err := policy.RunChecksForPolicy(attestationPolicy, &repoResult.RawResults)
	if err != nil {
		glog.Fatalf("Error when evaluating image %q against policy", image)
	}
	if result != policy.Pass {
		glog.Errorf("policy check failed on image %s:", image)
		os.Exit(1)
	}
	glog.Infof("Image %q passes policy check", image)
}
