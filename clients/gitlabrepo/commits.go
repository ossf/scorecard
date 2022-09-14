package gitlabrepo

import (
	"fmt"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type commitsHandler struct {
	glClient *gitlab.Client
	once     *sync.Once
	errSetup error
	repourl  *repoURL
	commits  []clients.Commit
}

func (handler *commitsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *commitsHandler) setup() error {
	handler.once.Do(func() {
		commits, _, err := handler.glClient.Commits.ListCommits(handler.repourl.projectID, &gitlab.ListCommitsOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("request for commits failed with %w", err)
			return
		}

		// To limit the number of user requests we are going to map every committer email
		// to a user.
		userToEmail := make(map[string]*gitlab.User)
		for _, commit := range commits {
			user, ok := userToEmail[commit.AuthorEmail]
			if !ok {
				users, _, err := handler.glClient.Search.Users(commit.CommitterName, &gitlab.SearchOptions{})
				if err != nil {
					// Possibility this shouldn't be an issue as individuals can leave organizations
					// (possibly taking their account with them)
					handler.errSetup = fmt.Errorf("unable to find user associated with commit: %w", err)
					return
				}
				userToEmail[commit.AuthorEmail] = users[0]
				user = users[0]
			}

			// Commits are able to be a part of multiple merge requests, but the only one that will be important
			// here is the earliest one.
			mergeRequests, _, err := handler.glClient.Commits.ListMergeRequestsByCommit(handler.repourl.projectID, commit.ID)
			if err != nil {
				// Possibly do not return here as newer commits may not be associated with merge requests
				// TODO: check out the above possibility
				handler.errSetup = fmt.Errorf("unable to find merge requests associated with commit: %w", err)
				return
			}
			// There has to be an argmin function I can use because this is probably super slow.
			// TODO: grab argmin implementation.
			var mergeRequest *gitlab.MergeRequest
			if len(mergeRequests) > 0 {
				idx := 0
				for i, mergeRequest := range mergeRequests {
					if mergeRequest.MergedAt.Before(*mergeRequests[idx].MergedAt) {
						idx = i
					}
				}
				mergeRequest = mergeRequests[idx]
			} else {
				handler.commits = append(handler.commits, clients.Commit{
					CommittedDate: *commit.CommittedDate,
					Message:       commit.Message,
					SHA:           commit.ID,
				})
				continue
			}

			// Casting the Reviewers into clients.Review.
			var reviews []clients.Review
			for _, reviewer := range mergeRequest.Reviewers {
				reviews = append(reviews, clients.Review{
					Author: &clients.User{ID: int64(reviewer.ID)},
					State:  "",
				})
			}

			// Casting the Labels into []clients.Label.
			var labels []clients.Label
			for _, label := range mergeRequest.Labels {
				labels = append(labels, clients.Label{
					Name: label,
				})
			}

			// append the commits to the handler.
			handler.commits = append(handler.commits,
				clients.Commit{
					CommittedDate: *commit.CommittedDate,
					Message:       commit.Message,
					SHA:           commit.ID,
					AssociatedMergeRequest: clients.PullRequest{
						Number:   mergeRequest.ID,
						MergedAt: *mergeRequest.MergedAt,
						HeadSHA:  mergeRequest.SHA,
						Author:   clients.User{ID: int64(mergeRequest.Author.ID)},
						Labels:   labels,
						Reviews:  reviews,
					},
					Committer: clients.User{ID: int64(user.ID)},
				})
		}
	})

	return handler.errSetup
}

func (handler *commitsHandler) listCommits() ([]clients.Commit, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	return handler.commits, nil
}
