// Copyright Security Scorecard Authors
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

package raw

import (
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

// Maintained checks for maintenance.
func Maintained(c *checker.CheckRequest) (checker.MaintainedData, error) {
	var result checker.MaintainedData

	// Archived status.
	archived, err := c.RepoClient.IsArchived()
	if err != nil {
		return result, fmt.Errorf("%w", err)
	}
	result.ArchivedStatus.Status = archived

	// Recent commits.
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		return result, fmt.Errorf("%w", err)
	}

	for i := range commits {
		// Note: getRawDataFromCommit() is defined in Code-Review check.
		result.DefaultBranchCommits = append(result.DefaultBranchCommits,
			getRawDataFromCommit(&commits[i]))
	}

	// Recent issues.
	issues, err := c.RepoClient.ListIssues()
	if err != nil {
		return result, fmt.Errorf("%w", err)
	}

	for i := range issues {
		// Create issue.
		issue := checker.Issue{
			CreatedAt: issues[i].CreatedAt,
		}
		// Add author if not nil.
		if issues[i].Author != nil {
			issue.Author = &checker.User{
				Login:           issues[i].Author.Login,
				RepoAssociation: getAssociation(issues[i].AuthorAssociation),
			}
		}
		// Add URL if not nil.
		if issues[i].URI != nil {
			issue.URL = *issues[i].URI
		}

		// Add comments.
		for j := range issues[i].Comments {
			comment := checker.Comment{
				CreatedAt: issues[i].Comments[j].CreatedAt,
			}
			if issues[i].Comments[j].Author != nil {
				comment.Author = &checker.User{
					Login:           issues[i].Comments[j].Author.Login,
					RepoAssociation: getAssociation(issues[i].Comments[j].AuthorAssociation),
				}
			}

			issue.Comments = append(issue.Comments, comment)
		}

		result.Issues = append(result.Issues, issue)
	}

	return result, nil
}

func getAssociation(a *clients.RepoAssociation) *checker.RepoAssociation {
	if a == nil {
		return nil
	}

	switch *a {
	case clients.RepoAssociationContributor:
		v := checker.RepoAssociationContributor
		return &v
	case clients.RepoAssociationCollaborator:
		v := checker.RepoAssociationCollaborator
		return &v
	case clients.RepoAssociationOwner:
		v := checker.RepoAssociationOwner
		return &v
	case clients.RepoAssociationMember:
		v := checker.RepoAssociationMember
		return &v
	case clients.RepoAssociationFirstTimer:
		v := checker.RepoAssociationFirstTimer
		return &v
	case clients.RepoAssociationMannequin:
		v := checker.RepoAssociationMannequin
		return &v
	case clients.RepoAssociationNone:
		v := checker.RepoAssociationNone
		return &v
	case clients.RepoAssociationFirstTimeContributor:
		v := checker.RepoAssociationFirstTimeContributor
		return &v
	default:
		return nil
	}
}
