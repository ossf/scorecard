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
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type licensesHandler struct {
	glProject *gitlab.Project
	once      *sync.Once
	errSetup  error
	repourl   *repoURL
	licenses  []clients.License
}

func (handler *licensesHandler) init(repourl *repoURL, project *gitlab.Project) {
	handler.repourl = repourl
	handler.glProject = project
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

var errLicenseURLParse = errors.New("couldn't parse gitlab repo license url")

func (handler *licensesHandler) setup() error {
	handler.once.Do(func() {
		l := handler.glProject.License

		ptn, err := regexp.Compile(fmt.Sprintf("%s/-/blob/(\\w+)/(.*)", handler.repourl.URI()))
		if err != nil {
			handler.errSetup = fmt.Errorf("couldn't parse license url: %w", err)
			return
		}

		m := ptn.FindStringSubmatch(handler.glProject.LicenseURL)
		if len(m) < 3 {
			handler.errSetup = fmt.Errorf("%w: %s", errLicenseURLParse, handler.glProject.LicenseURL)
			return
		}
		path := m[2]

		handler.licenses = append(handler.licenses,
			clients.License{
				Key:    l.Key,
				Name:   l.Name,
				Path:   path,
				SPDXId: l.Key,
			},
		)

		handler.errSetup = nil
	})

	return handler.errSetup
}

// Currently listLicenses() returns the percentages (truncated) of each language in the project.
func (handler *licensesHandler) listLicenses() ([]clients.License, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during licensesHandler.setup: %w", err)
	}

	return handler.licenses, nil
}
