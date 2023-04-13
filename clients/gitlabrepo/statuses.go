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

// for gitlab this only works if ref is SHA.
func (handler *statusesHandler) listStatuses(ref string) ([]clients.Status, error) {
	commitStatuses, _, err := handler.glClient.Commits.GetCommitStatuses(
		handler.repourl.project, ref, &gitlab.GetCommitStatusesOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting commit statuses: %w", err)
	}
	return statusFromData(commitStatuses), nil
}

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
