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
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/projectanalysis"

	"github.com/ossf/scorecard/v5/clients"
)

var errLanguageAnalyticsMalformed = errors.New("azure DevOps language analytics response is malformed")

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
	l.errSetup = nil
}

type (
	fnGetProjectLanguageAnalytics func(
		ctx context.Context,
		args projectanalysis.GetProjectLanguageAnalyticsArgs,
	) (*projectanalysis.ProjectLanguageAnalytics, error)
)

func isLanguageAnalyticsUnavailable(err error) bool {
	if err == nil {
		return false
	}

	var wrappedError azuredevops.WrappedError
	if errors.As(err, &wrappedError) && wrappedError.StatusCode != nil {
		switch *wrappedError.StatusCode {
		case 401, 403, 404:
			return true
		}
	}
	var wrappedErrorPointer *azuredevops.WrappedError
	if errors.As(err, &wrappedErrorPointer) && wrappedErrorPointer.StatusCode != nil {
		switch *wrappedErrorPointer.StatusCode {
		case 401, 403, 404:
			return true
		}
	}

	message := strings.ToLower(err.Error())
	for _, fragment := range []string{
		"access denied",
		"not authorized",
		"permission",
		"not available",
		"not enabled",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

func (l *languagesHandler) useAllLanguagesFallback() {
	l.languages = []clients.Language{{Name: clients.All, NumLines: 1}}
	l.errSetup = nil
}

func (l *languagesHandler) useProjectAnalytics(res *projectanalysis.ProjectLanguageAnalytics) error {
	if res.RepositoryLanguageAnalytics == nil {
		return fmt.Errorf("%w: missing repositories", errLanguageAnalyticsMalformed)
	}

	for _, repo := range *res.RepositoryLanguageAnalytics {
		if repo.Id == nil {
			return fmt.Errorf("%w: missing repository ID", errLanguageAnalyticsMalformed)
		}
		if repo.Id.String() != l.repourl.id {
			continue
		}
		if repo.LanguageBreakdown == nil {
			return fmt.Errorf("%w: missing language breakdown", errLanguageAnalyticsMalformed)
		}

		// TODO: Find the number of lines in the repo and multiply the value of each language by that number.
		for _, language := range *repo.LanguageBreakdown {
			if language.Name == nil {
				return fmt.Errorf("%w: missing language name", errLanguageAnalyticsMalformed)
			}
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
	return nil
}

func (l *languagesHandler) setup() error {
	l.once.Do(func() {
		args := projectanalysis.GetProjectLanguageAnalyticsArgs{
			Project: &l.repourl.project,
		}
		res, err := l.projectAnalysis(l.ctx, args)
		if err != nil {
			if isLanguageAnalyticsUnavailable(err) {
				l.useAllLanguagesFallback()
				return
			}
			l.errSetup = err
			return
		}

		if res == nil {
			l.errSetup = fmt.Errorf("%w: missing response", errLanguageAnalyticsMalformed)
			return
		}
		if res.ResultPhase == nil || *res.ResultPhase != projectanalysis.ResultPhaseValues.Full {
			l.useAllLanguagesFallback()
			return
		}
		l.errSetup = l.useProjectAnalytics(res)
	})
	return l.errSetup
}

func (l *languagesHandler) listProgrammingLanguages() ([]clients.Language, error) {
	if err := l.setup(); err != nil {
		return nil, err
	}

	return l.languages, nil
}
