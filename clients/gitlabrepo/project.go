package gitlabrepo

import (
	"fmt"
	"sync"
	"time"

	"github.com/xanzy/go-gitlab"
)

type projectHandler struct {
	glClient  *gitlab.Client
	once      *sync.Once
	errSetup  error
	repourl   *repoURL
	createdAt time.Time
	archived  bool
}

func (handler *projectHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *projectHandler) setup() error {
	handler.once.Do(func() {
		proj, _, err := handler.glClient.Projects.GetProject(handler.repourl.projectID, &gitlab.GetProjectOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("request for project failed with error %w", err)
			return
		}

		handler.createdAt = *proj.CreatedAt
		handler.archived = proj.Archived
	})

	return handler.errSetup
}

func (handler *projectHandler) isArchived() (bool, error) {
	if err := handler.setup(); err != nil {
		return true, fmt.Errorf("error during projectHandler.setup: %w", err)
	}

	return handler.archived, nil
}

func (handler *projectHandler) getCreatedAt() (time.Time, error) {
	if err := handler.setup(); err != nil {
		return time.Now(), fmt.Errorf("error during projectHandler.setup: %w", err)
	}

	return handler.createdAt, nil
}
