package gitlabrepo

import (
	"fmt"
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
		issues, _, err := handler.glClient.Issues.ListProjectIssues(handler.repourl.projectID, &gitlab.ListProjectIssuesOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("unable to find issues associated with the project id: %w", err)
			return
		}

		if len(issues) > 0 {
			for _, issue := range issues {
				handler.issues = append(handler.issues,
					clients.Issue{
						URI:       &issue.ExternalID,
						CreatedAt: issue.CreatedAt,
						Author: &clients.User{
							ID: int64(issue.Author.ID),
						},
						Comments: nil,
					})
			}
		} else {
			handler.issues = nil
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
