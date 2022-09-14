package gitlabrepo

import (
	"fmt"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type languagesHandler struct {
	glClient  *gitlab.Client
	once      *sync.Once
	errSetup  error
	repourl   *repoURL
	languages []clients.Language
}

func (handler *languagesHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *languagesHandler) setup() error {
	handler.once.Do(func() {
		client := handler.glClient
		languageMap, _, err := client.Projects.GetProjectLanguages(handler.repourl.projectID)
		if err != nil || languageMap == nil {
			handler.errSetup = fmt.Errorf("request for repo languages failed with %w", err)
			return
		}

		// TODO: find how to find number of lines in a gitlab project.
		const placeholder = 100000

		for k, v := range *languageMap {
			handler.languages = append(handler.languages,
				clients.Language{
					Name:     clients.LanguageName(k),
					NumLines: int(v * placeholder),
				},
			)
		}
		handler.errSetup = nil
	})

	return handler.errSetup
}

func (handler *languagesHandler) listProgrammingLanguages() ([]clients.Language, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during languagesHandler.setup: %w", err)
	}

	return handler.languages, nil
}
