// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"context"
	"log"
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/projectanalysis"

	"github.com/ossf/scorecard/v5/clients"
)

type languagesHandler struct {
	ctx                   context.Context
	once                  *sync.Once
	repourl               *Repo
	projectAnalysisClient projectanalysis.Client
	projectAnalysis       fnGetProjectLanguageAnalytics
	errSetup              error
	languages             []clients.Language
}

func (l *languagesHandler) init(ctx context.Context, repourl *Repo) {
	l.ctx = ctx
	l.once = new(sync.Once)
	l.repourl = repourl
	l.languages = []clients.Language{}
	l.projectAnalysis = l.projectAnalysisClient.GetProjectLanguageAnalytics
}

type (
	fnGetProjectLanguageAnalytics func(
		ctx context.Context,
		args projectanalysis.GetProjectLanguageAnalyticsArgs,
	) (*projectanalysis.ProjectLanguageAnalytics, error)
)
func (l *languagesHandler) setup() error {
	l.once.Do(func() {
		args := projectanalysis.GetProjectLanguageAnalyticsArgs{
			Project: &l.repourl.project,
		}
		res, err := l.projectAnalysis(l.ctx, args)
		if err != nil {
			l.errSetup = err
			return
		}

		if res.ResultPhase != &projectanalysis.ResultPhaseValues.Full {
			log.Println("Project language analytics not ready yet. Results may be incomplete.")
		}

		for _, repo := range *res.RepositoryLanguageAnalytics {
			if repo.Id.String() != l.repourl.id {
				continue
			}

			// TODO: Find the number of lines in the repo and multiply the value of each language by that number.
			for _, language := range *repo.LanguageBreakdown {
				percentage := 0
				if language.LanguagePercentage != nil {
					percentage = int(*language.LanguagePercentage)
				}
				l.languages = append(l.languages,
					clients.Language{
						Name:     clients.LanguageName(*language.Name),
						NumLines: percentage,
					},
				)
			}

		}
	})
	return l.errSetup
}

func (l *languagesHandler) listProgrammingLanguages() ([]clients.Language, error) {
	if err := l.setup(); err != nil {
		return nil, err
	}

	return l.languages, nil
}
