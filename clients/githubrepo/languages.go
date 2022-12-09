// Copyright 2021 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
)

type languagesHandler struct {
	ghclient  *github.Client
	once      *sync.Once
	ctx       context.Context
	errSetup  error
	repourl   *repoURL
	languages []clients.Language
}

func (handler *languagesHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.languages = nil
}

// TODO: Can add support to parse the raw response JSON and mark languages that are not in
// our defined Language consts in clients/languages.go as "not supported languages".
func (handler *languagesHandler) setup() error {
	handler.once.Do(func() {
		client := handler.ghclient
		reqURL := path.Join("repos", handler.repourl.owner, handler.repourl.repo, "languages")
		req, err := client.NewRequest("GET", reqURL, nil)
		if err != nil {
			handler.errSetup = fmt.Errorf("request for repo languages failed with %w", err)
			return
		}
		bodyJSON := map[clients.LanguageName]int{}
		// The client.repoClient.Do API writes the response body to var bodyJSON,
		// so we can ignore the first returned variable (the entire http response object)
		// since we only need the response body here.
		_, err = client.Do(handler.ctx, req, &bodyJSON)
		if err != nil {
			handler.errSetup = fmt.Errorf("response for repo languages failed with %w", err)
			return
		}
		// Parse the raw JSON to an array of languages.
		for k, v := range bodyJSON {
			// TODO: once the const defined in clients/languages.go becomes a complete list of langs supported,
			// add support here so that for not supported langs, it emits an "not-supported" error and break the parse.
			// Currently, we are parsing all the JSON-returned langs into the result since the const is incomplete.
			handler.languages = append(handler.languages,
				clients.Language{
					Name:     k,
					NumLines: v,
				},
			)
		}
		handler.errSetup = nil
	})

	return handler.errSetup
}

func (handler *languagesHandler) listProgrammingLanguages() ([]clients.Language, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during languagesHandler.setup: %w", err)
	}
	return handler.languages, nil
}
