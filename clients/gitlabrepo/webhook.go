package gitlabrepo

import (
	"fmt"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type webhookHandler struct {
	glClient *gitlab.Client
	once     *sync.Once
	errSetup error
	repourl  *repoURL
	webhooks []clients.Webhook
}

func (handler *webhookHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *webhookHandler) setup() error {
	handler.once.Do(func() {
		projectHooks, _, err := handler.glClient.Projects.ListProjectHooks(
			handler.repourl.projectID, &gitlab.ListProjectHooksOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("request for project hooks failed with %w", err)
			return
		}

		// TODO: make sure that enablesslverification is similarly equivalent to auth secret.
		for _, hook := range projectHooks {
			handler.webhooks = append(handler.webhooks,
				clients.Webhook{
					Path:           hook.URL,
					ID:             int64(hook.ID),
					UsesAuthSecret: hook.EnableSSLVerification,
				})
		}
	})

	return handler.errSetup
}

func (handler *webhookHandler) listWebhooks() ([]clients.Webhook, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during webhookHandler.setup: %w", err)
	}

	return handler.webhooks, nil
}
