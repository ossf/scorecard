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
	"strings"
	"sync"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

// issuesHandler fetches and caches basic issue metadata for a project.
type issuesHandler struct {
	errSetup        error
	glClient        *gitlab.Client
	repourl         *Repo
	once            *sync.Once
	maintainers     map[string]struct{}
	issues          []clients.Issue
	maintainersOnce sync.Once
}

func (h *issuesHandler) init(repourl *Repo) {
	h.repourl = repourl
	h.errSetup = nil
	h.once = new(sync.Once)
}

// setup collects a minimal list of issues (URI, IssueNumber, CreatedAt, ClosedAt).
//
//nolint:gocognit // Complex GitLab API pagination with maintainer detection logic
func (h *issuesHandler) setup() error {
	h.once.Do(func() {
		project := h.repourl.projectID // field, not method

		// There doesn't seem to be a good way to get user access_levels in gitlab so the following way may seem incredibly
		// barbaric, however I couldn't find a better way in the docs.
		projMemberships, resp, err := h.glClient.ProjectMembers.ListAllProjectMembers(
			project, &gitlab.ListProjectMembersOptions{})
		if resp == nil {
			h.errSetup = fmt.Errorf("unable to find access tokens associated with the project id: %w", err)
			return
		}
		if err != nil && resp.StatusCode != http.StatusUnauthorized {
			h.errSetup = fmt.Errorf("unable to find access tokens associated with the project id: %w", err)
			return
		} else if resp.StatusCode == http.StatusUnauthorized {
			h.errSetup = fmt.Errorf("insufficient permissions to check issue author associations %w", err)
			return
		}

		page := 1
		var all []clients.Issue

		for {
			state := "all"
			issues, resp, err := h.glClient.Issues.ListProjectIssues(project, &gitlab.ListProjectIssuesOptions{
				State:       &state,
				ListOptions: gitlab.ListOptions{PerPage: 100, Page: page},
			})
			if err != nil {
				h.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("gitlab: ListProjectIssues: %v", err))
				return
			}

			for _, is := range issues {
				// URI must be the issue web URL
				url := is.WebURL

				// Map GitLab author → Scorecard User
				var author *clients.User
				if is.Author != nil {
					author = &clients.User{ID: int64(is.Author.ID)}
				}

				// Determine author association via project membership lookup
				authorAssociation := clients.RepoAssociationNone
				for _, m := range projMemberships {
					if is.Author != nil && is.Author.ID == m.ID {
						authorAssociation = accessLevelToRepoAssociation(m.AccessLevel)
					}
				}

				ci := clients.Issue{
					URI:               &url,
					CreatedAt:         is.CreatedAt, // *time.Time
					ClosedAt:          is.ClosedAt,  // *time.Time
					Author:            author,       // *clients.User
					AuthorAssociation: &authorAssociation,
					IssueNumber:       0, // required for Test_listIssues
					// Note: Labels, Comments, LabelEvents are added in getIssuesWithHistory
				}

				all = append(all, ci)
			}

			if resp == nil || resp.CurrentPage >= resp.TotalPages {
				break
			}
			page++
		}

		h.issues = all
	})
	return h.errSetup
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

func (h *issuesHandler) getIssues() ([]clients.Issue, error) {
	if err := h.setup(); err != nil {
		return nil, fmt.Errorf("issuesHandler.setup: %w", err)
	}
	return h.issues, nil
}

func (h *issuesHandler) listIssues() ([]clients.Issue, error) {
	if err := h.setup(); err != nil {
		return nil, fmt.Errorf("error during issuesHandler.setup: %w", err)
	}

	return h.issues, nil
}

// ---- enrichment for the new check ----

func (h *issuesHandler) ensureMaintainers() {
	h.maintainersOnce.Do(func() {
		h.maintainers = map[string]struct{}{}
		project := h.repourl.projectID

		page := 1
		for {
			members, resp, err := h.glClient.ProjectMembers.ListAllProjectMembers(
				project,
				&gitlab.ListProjectMembersOptions{
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: page},
				},
			)
			if err != nil || members == nil {
				return
			}
			for _, m := range members {
				if m.AccessLevel >= gitlab.DeveloperPermissions {
					login := strings.ToLower(m.Username)
					if login != "" {
						h.maintainers[login] = struct{}{}
					}
				}
			}
			if resp == nil || resp.CurrentPage >= resp.TotalPages {
				break
			}
			page++
		}
	})
}

//nolint:gocognit // Complex issue enrichment with label event and comment processing
func (h *issuesHandler) getIssuesWithHistory() ([]clients.Issue, error) {
	if err := h.setup(); err != nil {
		return nil, fmt.Errorf("issuesHandler.setup: %w", err)
	}
	h.ensureMaintainers()
	project := h.repourl.projectID

	out := make([]clients.Issue, 0, len(h.issues))
	for i := range h.issues {
		ci := h.issues[i] // copy

		// Label events (bug/security)
		lePage := 1
		for {
			evs, resp, err := h.glClient.ResourceLabelEvents.ListIssueLabelEvents(
				project,
				ci.IssueNumber,
				&gitlab.ListLabelEventsOptions{
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: lePage},
				},
			)
			if err != nil {
				return nil, sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("gitlab: ListIssueLabelEvents(issue #%d): %v", ci.IssueNumber, err))
			}
			for _, ev := range evs {
				// ev.Label is a struct (not a pointer)
				if ev.Label.Name == "" || ev.CreatedAt == nil {
					continue
				}
				name := strings.ToLower(ev.Label.Name)
				if name != "bug" && name != "security" {
					continue
				}
				switch ev.Action {
				case "add", "added":
					ci.LabelEvents = append(ci.LabelEvents, clients.LabelEvent{
						Label:     name,
						Added:     true,
						Actor:     "",
						CreatedAt: *ev.CreatedAt,
					})
				case "remove", "removed":
					ci.LabelEvents = append(ci.LabelEvents, clients.LabelEvent{
						Label:     name,
						Added:     false,
						Actor:     "",
						CreatedAt: *ev.CreatedAt,
					})
				}
			}
			if resp == nil || resp.CurrentPage >= resp.TotalPages {
				break
			}
			lePage++
		}

		// Comments (notes)
		nPage := 1
		for {
			notes, resp, err := h.glClient.Notes.ListIssueNotes(
				project,
				ci.IssueNumber,
				&gitlab.ListIssueNotesOptions{
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: nPage},
				},
			)
			if err != nil {
				return nil, sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("gitlab: ListIssueNotes(issue #%d): %v", ci.IssueNumber, err))
			}
			for _, n := range notes {
				login := strings.ToLower(n.Author.Username)
				_, isMaint := h.maintainers[login]

				created := time.Time{}
				if n.CreatedAt != nil {
					created = *n.CreatedAt
				}

				// GitLab Note doesn't expose a direct WebURL via this client.
				// Fall back to the issue URI if present.
				commentURL := ""
				if ci.URI != nil {
					commentURL = *ci.URI
				}

				ci.Comments = append(ci.Comments, clients.IssueComment{
					Author:       &clients.User{Login: login},
					CreatedAt:    &created,
					IsMaintainer: isMaint,
					URL:          commentURL,
				})
			}
			if resp == nil || resp.CurrentPage >= resp.TotalPages {
				break
			}
			nPage++
		}

		out = append(out, ci)
	}

	return out, nil
}
