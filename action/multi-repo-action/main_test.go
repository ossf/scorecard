// Copyright 2022 OpenSSF Authors
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
// SPDX-License-Identifier: Apache-2.0

//nolint
// TODO(lint): Remove nolint directive and fix lint warnings
package main

import (
	"github.com/google/go-github/v42/github"
)

var client *github.Client

// Currently incomplete
//nolint:lll
// Good reference: https://github.com/google/go-github/blob/887f605dd1f81715a4d4e3983e38450b29833639/github/repos_contents_test.go
// Currently from: https://github.com/google/go-github/blob/master/test/integration/repos_test.go

// TODO: Add/refactor tests
/*
func Test_OrgWorkflowAdd(t *testing.T) {
	client = github.NewClient(nil)
	me, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		t.Fatalf("Users.Get('') returned error: %v", err)
	}

	repo, err := createRandomTestRepository(*me.Login, false)
	if err != nil {
		t.Fatalf("createRandomTestRepository returned error: %v", err)
	}

	// update the repository description
	repo.Description = github.String("description")
	repo.DefaultBranch = nil // FIXME: this shouldn't be necessary
	_, _, err = client.Repositories.Edit(context.Background(), *repo.Owner.Login, *repo.Name, repo)
	if err != nil {
		t.Fatalf("Repositories.Edit() returned error: %v", err)
	}

	// delete the repository
	_, err = client.Repositories.Delete(context.Background(), *repo.Owner.Login, *repo.Name)
	if err != nil {
		t.Fatalf("Repositories.Delete() returned error: %v", err)
	}

	// verify that the repository was deleted
	_, resp, err := client.Repositories.Get(context.Background(), *repo.Owner.Login, *repo.Name)
	if err == nil {
		t.Fatalf("Test repository still exists after deleting it.")
	}
	if err != nil && resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Repositories.Get() returned error: %v", err)
	}
}

func createRandomTestRepository(owner string, autoinit bool) (*github.Repository, error) {
	// create random repo name that does not currently exist
	var repoName string
	for {
		repoName = fmt.Sprintf("test-1")
		_, resp, err := client.Repositories.Get(context.Background(), owner, repoName)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				// found a non-existent repo, perfect
				break
			}

			return nil, err
		}
	}

	// create the repository
	repo, _, err := client.Repositories.Create(
		context.Background(),
		"",
		&github.Repository{
			Name:     github.String(repoName),
			AutoInit: github.Bool(autoinit),
		},
	)
	if err != nil {
		return nil, err
	}

	return repo, nil
}
*/
