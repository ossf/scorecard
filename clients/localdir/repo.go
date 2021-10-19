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
	"os"
	"path"

	clients "github.com/ossf/scorecard/v3/clients"
)

type repoLocal struct {
	path string
}

// URI implements Repo.URI().
func (r *repoLocal) URI() string {
	return fmt.Sprintf("file://%s", r.Path())
}

func (r *repoLocal) Path() string {
	return r.path
}

// String implements Repo.String.
func (r *repoLocal) String() string {
	return r.URI()
}

// Org implements Repo.Org.
func (r *repoLocal) Org() clients.Repo {
	// TODO
	panic("Org")
	return &repoLocal{}
}

// IsValid implements Repo.IsValid.
func (r *repoLocal) IsValid() error {
	_, err := os.Stat(r.path)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// Metadata implements Repo.Metadata.
func (r *repoLocal) Metadata() []string {
	// TODO
	return []string{}
}

// AppendMetadata implements Repo.AppendMetadata.
func (r *repoLocal) AppendMetadata(m ...string) {
	// TODO
	panic("append meta")
}

// IsScorecardRepo implements Repo.IsScorecardRepo.
func (r *repoLocal) IsScorecardRepo() bool {
	// TODO
	panic("is scorecard repo")
	return false
}

// MakeLocalDirRepo returns an implementation of clients.Repo interface.
func MakeLocalDirRepo(pathfn string) (clients.Repo, error) {
	p := path.Clean(pathfn)
	repo := &repoLocal{
		path: p,
	}

	if err := repo.IsValid(); err != nil {
		return nil, fmt.Errorf("error in IsValid: %w", err)
	}

	return repo, nil
}
