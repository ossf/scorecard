// Copyright 2021 Security Scorecard Authors
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
//

// Package localdir is local repo containing source code.
package localdir

import (
	"fmt"

	clients "github.com/ossf/scorecard/v2/clients"
)

type localRepoClient struct {
	path string
}

func (r *localRepoClient) URI() string {
	return fmt.Sprintf("file://%s", r.path)
}

func (r *localRepoClient) String() string {
	panic("invalid String()")
	//nolint
	return ""
}

func (r *localRepoClient) Org() clients.Repo {
	panic("invalid Org()")
	//nolint
	return &localRepoClient{}
}

func (r *localRepoClient) IsValid() error {
	panic("invalid IsValid()")
	//nolint
	return nil
}

func (r *localRepoClient) Metadata() []string {
	panic("invalid Metadata()")
	//nolint
	return nil
}

func (r *localRepoClient) AppendMetadata(m ...string) {
	panic("invalid AppendMetadata()")
}

func (r *localRepoClient) IsScorecardRepo() bool {
	panic("invalid IsScorecardRepo()")
	//nolint
	return false
}

// MakeLocalDirRepo returns an implementation of clients.Repo interface.
func MakeLocalDirRepo(path string) (clients.Repo, error) {
	// TODO: validate the path exists
	return &localRepoClient{
		path: path,
	}, nil
}
