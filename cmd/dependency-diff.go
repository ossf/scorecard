// Copyright 2022 Security Scorecard Authors
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

// Package cmd implements Scorecard commandline.
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/dependencydiff"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/options"
	"github.com/ossf/scorecard/v4/pkg"
)

const (
	dependencydiffUse   = `dependency-diff --commit=<base>...<head> --repo=<repo> [--checks=check1,...]`
	dependencydiffShort = "Surface Scorecard checking results on dependency-diffs between commits or branches"
)

func dependencydiffCmd(o *options.Options) *cobra.Command {
	depdiffCmd := &cobra.Command{
		Use:   dependencydiffUse,
		Short: dependencydiffShort,
		Long:  ``,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := o.ValidateDepdiff()
			if err != nil {
				return fmt.Errorf("validating options: %w", err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			logger := sclog.NewLogger(sclog.ParseLevel(o.LogLevel))
			checkDocs, err := docs.Read()
			if err != nil {
				log.Panicf("cannot read yaml file: %v", err)
			}
			doDependencydiff(ctx, o, logger, checkDocs)
		},
	}
	return depdiffCmd
}

func doDependencydiff(ctx context.Context, o *options.Options,
	logger *sclog.Logger, checkDocs docs.Doc,
) {
	commits := strings.Split(o.Commit, "...")
	if len(commits) != 2 {
		log.Panicf("error in commits: %v", os.ErrInvalid)
	}
	base, head := commits[0], commits[1]
	ownerRepo := strings.Split(o.Repo, "/")
	if len(ownerRepo) != 2 {
		log.Panicf("error in repo: %v", os.ErrInvalid)
	}
	owner, repo := ownerRepo[0], ownerRepo[1]
	depdiffResults, err := dependencydiff.GetDependencyDiffResults(
		ctx, logger, owner, repo, base, head, o.ChecksToRun, nil)
	if err != nil {
		log.Panicf("error getting dependencydiff results: %v", err)
	}
	depdiffErr := pkg.FormatDependencydiffResults(o, depdiffResults, checkDocs)
	if depdiffErr != nil {
		log.Panicf("Failed to format dependencydiff results: %v", depdiffErr)
	}
}
