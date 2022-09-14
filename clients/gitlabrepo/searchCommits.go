package gitlabrepo

import (
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type searchCommitsHandler struct {
	glClient *gitlab.Client
	repourl  *repoURL
}

func (handler *searchCommitsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
}

func (handler *searchCommitsHandler) search(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return nil, fmt.Errorf("%w: Search only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	query, err := handler.buildQuery(request)
	if err != nil {
		return nil, fmt.Errorf("handler.buildQuiery: %w", err)
	}

	// TODO: I'm skeptical this will work as intended
	commits, _, err := handler.glClient.Search.CommitsByProject(handler.repourl.projectID, query, &gitlab.SearchOptions{})
	if err != nil {
		return nil, fmt.Errorf("Search.Commits: %w", err)
	}

	// Gitlab returns a list of commits that does not contain the committer's id, unlike in
	// githubrepo/searchCommits.go so to limit the number of requests we are mapping each unique user
	// email to thei gitlab user data.
	userMap := make(map[string]*gitlab.User)
	var ret []clients.Commit
	for _, commit := range commits {
		if _, ok := userMap[commit.CommitterEmail]; !ok {
			user, _, err := handler.glClient.Search.Users(commit.CommitterEmail, &gitlab.SearchOptions{})
			if err != nil {
				return nil, fmt.Errorf("gitlab-searchCommits: %w", err)
			}
			userMap[commit.CommitterEmail] = user[0]
		}

		ret = append(ret, clients.Commit{
			Committer: clients.User{ID: int64(userMap[commit.CommitterEmail].ID)},
		})
	}

	return ret, nil
}

func (handler *searchCommitsHandler) buildQuery(request clients.SearchCommitsOptions) (string, error) {
	if request.Author == "" {
		return "", fmt.Errorf("%w", errEmptyQuery)
	}
	var queryBuilder strings.Builder
	if _, err := queryBuilder.WriteString(
		fmt.Sprintf("repo:%s/%s author:%s",
			handler.repourl.owner, handler.repourl.projectID,
			request.Author)); err != nil {
		return "", fmt.Errorf("writestring: %w", err)
	}

	return queryBuilder.String(), nil
}
