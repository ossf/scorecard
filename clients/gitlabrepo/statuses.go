package gitlabrepo

import (
	"fmt"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type statusesHandler struct {
	glClient *gitlab.Client
	repourl  *repoURL
}

func (handler *statusesHandler) init(repourl *repoURL) {
	handler.repourl = repourl
}

// The docs were a bit fuzzy, but I'm pretty sure this is commit statuses
// for gitlab this only works if ref is a SHA so I'll add that to the docs.
func (handler *statusesHandler) listStatuses(ref string) ([]clients.Status, error) {
	commitStatuses, _, err := handler.glClient.Commits.GetCommitStatuses(
		handler.repourl.projectID, ref, &gitlab.GetCommitStatusesOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting commit statuses: %w", err)
	}
	return statusFromData(commitStatuses), nil
}

// I'm not 100% sure what the difference is between URL and targetURL.
func statusFromData(commitStatuses []*gitlab.CommitStatus) []clients.Status {
	var statuses []clients.Status
	for _, commitStatus := range commitStatuses {
		statuses = append(statuses, clients.Status{
			State:     commitStatus.Status,
			Context:   fmt.Sprint(commitStatus.ID),
			URL:       commitStatus.TargetURL,
			TargetURL: commitStatus.TargetURL,
		})
	}
	return statuses
}
