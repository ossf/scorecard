package gitlabrepo

import (
	"fmt"
	"strings"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type contributorsHandler struct {
	glClient     *gitlab.Client
	once         *sync.Once
	errSetup     error
	repourl      *repoURL
	contributors []clients.User
}

func (handler *contributorsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *contributorsHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListContributors only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}
		contribs, _, err := handler.glClient.Repositories.Contributors(handler.repourl.projectID, &gitlab.ListContributorsOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("error during ListContributors: %w", err)
			return
		}

		for _, contrib := range contribs {
			if contrib.Name == "" {
				continue
			}

			// In Gitlab users only have one registered organization which is the company they work for, this means that
			// the organizations field will not be filled in and the companies field will be a singular value.
			users, _, err := handler.glClient.Search.Users(contrib.Name, &gitlab.SearchOptions{})
			if err != nil {
				handler.errSetup = fmt.Errorf("error during Users.Get: %w", err)
				return
			} else if len(users) == 0 {
				// parseEmailToName is declared in commits.go
				users, _, err = handler.glClient.Search.Users(parseEmailToName(contrib.Email), &gitlab.SearchOptions{})
				if err != nil {
					handler.errSetup = fmt.Errorf("error during Users.Get: %w", err)
					return
				}
			}

			contributor := clients.User{
				Login:            contrib.Email,
				Companies:        []string{users[0].Organization},
				NumContributions: contrib.Commits,
				ID:               int64(users[0].ID),
			}
			handler.contributors = append(handler.contributors, contributor)
		}
	})
	return handler.errSetup
}

func (handler *contributorsHandler) getContributors() ([]clients.User, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during contributorsHandler.setup: %w", err)
	}
	return handler.contributors, nil
}
