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

package gitlabrepo

// TODO:
// add "github.com/xanzy/go-gitlab" to this list.

import (
	"fmt"
	"sync"

	"github.com/ossf/scorecard/v4/clients"
)

type licensesHandler struct {
	// TODO: glClient *gitlab.Client
	once     *sync.Once
	errSetup error
	repourl  *repoURL
	licenses []clients.License
}

func (handler *licensesHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *licensesHandler) setup() error {
	handler.once.Do(func() {
		// TODO: find actual GitLab API, data type, and fields
		// client := handler.glClient
		// licenseMap, _, err := client.Projects.GetLicense(handler.repourl.projectID)
		licenseMap := []clients.License{}
		// TODO: err := (*struct{})(nil)
		if len(licenseMap) == 0 {
			// TODO: handler.errSetup = fmt.Errorf("request for repo licenses failed with %w", err)
			handler.errSetup = fmt.Errorf("%w: ListLicenses not yet supported for gitlab", clients.ErrUnsupportedFeature)
			return
		}

		// TODO: find actual GitLab API, data type, and fields
		// TODO: for k, v := range *licenseMap {
		//		handler.licenses = append(handler.licenses,
		//			clients.License{
		//				Key:    "",
		//				Name:   "",
		//				Path:   "",
		//				Size:   0,
		//				SPDXId: "",
		//				Type:   "",
		//			},
		//		)
		//	}
		//
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
