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
	"net/http"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type issuesHandler struct {
	glClient *gitlab.Client
	once     *sync.Once
	errSetup error
	repourl  *repoURL
	issues   []clients.Issue
}

func (handler *issuesHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *issuesHandler) setup() error {
	handler.once.Do(func() {
		issues, _, err := handler.glClient.Issues.ListProjectIssues(
			handler.repourl.projectID, &gitlab.ListProjectIssuesOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("unable to find issues associated with the project id: %w", err)
			return
		}

		// There doesn't seem to be a good way to get user access_levels in gitlab so the following way may seem incredibly
		// barberic, however I couldn't find a better way in the docs.
		projMemberships, resp, err := handler.glClient.ProjectMembers.ListAllProjectMembers(
			handler.repourl.projectID, &gitlab.ListProjectMembersOptions{})
		if err != nil && resp.StatusCode != 401 {
			handler.errSetup = fmt.Errorf("unable to find access tokens associated with the project id: %w", err)
			return
		} else if resp.StatusCode == http.StatusUnauthorized {
			handler.errSetup = fmt.Errorf("insufficient permissions to check issue author associations %w", err)
			return
		}

		var authorAssociation clients.RepoAssociation
		for _, issue := range issues {
			for _, m := range projMemberships {
				if issue.Author.ID == m.ID {
					authorAssociation = accessLevelToRepoAssociation(m.AccessLevel)
				}
			}

			issueIDString := fmt.Sprint(issue.ID)
			handler.issues = append(handler.issues,
				clients.Issue{
					URI:       &issueIDString,
					CreatedAt: issue.CreatedAt,
					Author: &clients.User{
						ID: int64(issue.Author.ID),
					},
					AuthorAssociation: &authorAssociation,
					Comments:          nil,
				})
		}
	})
	return handler.errSetup
}

func (handler *issuesHandler) listIssues() ([]clients.Issue, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during issuesHandler.setup: %w", err)
	}

	return handler.issues, nil
}

func accessLevelToRepoAssociation(l gitlab.AccessLevelValue) clients.RepoAssociation {
	switch l {
	case 0:
		return clients.RepoAssociationNone
	case 5:
		return clients.RepoAssociationFirstTimeContributor
	case 10:
		return clients.RepoAssociationCollaborator
	case 20:
		return clients.RepoAssociationCollaborator
	case 30:
		return clients.RepoAssociationMember
	case 40:
		return clients.RepoAssociationMaintainer
	case 50:
		return clients.RepoAssociationOwner
	default:
		return clients.RepoAssociationNone
	}
}
