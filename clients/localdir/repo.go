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
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	clients "github.com/ossf/scorecard/v3/clients"
)

var (
	errNotDirectory = errors.New("not a directory")
	errInvalidURI   = errors.New("invalid URI")
)

var filePrefix = "file://"

type repoLocal struct {
	path     string
	metadata []string
}

// URI implements Repo.URI().
func (r *repoLocal) URI() string {
	return fmt.Sprintf("file://%s", r.path)
}

// String implements Repo.String.
func (r *repoLocal) String() string {
	return r.URI()
}

// Org implements Repo.Org.
func (r *repoLocal) Org() clients.Repo {
	return nil
}

// IsValid implements Repo.IsValid.
func (r *repoLocal) IsValid() error {
	f, err := os.Stat(r.path)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if !f.IsDir() {
		return fmt.Errorf("%w", errNotDirectory)
	}
	return nil
}

// Metadata implements Repo.Metadata.
func (r *repoLocal) Metadata() []string {
	return []string{}
}

// AppendMetadata implements Repo.AppendMetadata.
func (r *repoLocal) AppendMetadata(m ...string) {
	r.metadata = append(r.metadata, m...)
}

// MakeLocalDirRepo returns an implementation of clients.Repo interface.
func MakeLocalDirRepo(pathfn string) (clients.Repo, error) {
	if !strings.HasPrefix(pathfn, filePrefix) {
		return nil, fmt.Errorf("%w", errInvalidURI)
	}
	p := path.Clean(pathfn[len(filePrefix):])
	repo := &repoLocal{
		path: p,
	}

	if err := repo.IsValid(); err != nil {
		return nil, fmt.Errorf("error in IsValid: %w", err)
	}
	return repo, nil
}
