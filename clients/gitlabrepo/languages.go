// Copyright 2022 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"fmt"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type languagesHandler struct {
	glClient  *gitlab.Client
	once      *sync.Once
	errSetup  error
	repourl   *repoURL
	languages []clients.Language
}

func (handler *languagesHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *languagesHandler) setup() error {
	handler.once.Do(func() {
		client := handler.glClient
		languageMap, _, err := client.Projects.GetProjectLanguages(handler.repourl.projectID)
		if err != nil || languageMap == nil {
			handler.errSetup = fmt.Errorf("request for repo languages failed with %w", err)
			return
		}

		// TODO(#2266): find number of lines of gitlab project and multiple the value of each language by that number.
		for k, v := range *languageMap {
			handler.languages = append(handler.languages,
				clients.Language{
					Name:     clients.LanguageName(k),
					NumLines: int(v * 100),
				},
			)
		}
		handler.errSetup = nil
	})

	return handler.errSetup
}

// Currently listProgrammingLanguages() returns the percentages (truncated) of each language in the project.
func (handler *languagesHandler) listProgrammingLanguages() ([]clients.Language, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during languagesHandler.setup: %w", err)
	}

	return handler.languages, nil
}
