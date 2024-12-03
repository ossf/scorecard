// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/audit"
)

type auditHandler struct {
	auditClient audit.Client
	once        *sync.Once
	ctx         context.Context
	errSetup    error
	repourl     *Repo
	createdAt   time.Time
	queryLog    fnQueryLog
}

func (a *auditHandler) init(ctx context.Context, repourl *Repo) {
	a.ctx = ctx
	a.errSetup = nil
	a.once = new(sync.Once)
	a.repourl = repourl
	a.queryLog = a.auditClient.QueryLog
}

type (
	fnQueryLog func(ctx context.Context, args audit.QueryLogArgs) (*audit.AuditLogQueryResult, error)
)

func (a *auditHandler) setup() error {
	a.once.Do(func() {
		continuationToken := ""
		for {
			auditLog, err := a.queryLog(a.ctx, audit.QueryLogArgs{
				ContinuationToken: &continuationToken,
			})
			if err != nil {
				a.errSetup = fmt.Errorf("error querying audit log: %w", err)
				return
			}

			// Check if Git.CreateRepo event exists for the repository
			for i := range *auditLog.DecoratedAuditLogEntries {
				entry := &(*auditLog.DecoratedAuditLogEntries)[i]
				if *entry.ActionId == "Git.CreateRepo" &&
					*entry.ProjectName == a.repourl.project &&
					(*entry.Data)["RepoName"] == a.repourl.name {
					a.createdAt = entry.Timestamp.Time
					return
				}
			}

			if *auditLog.HasMore {
				continuationToken = *auditLog.ContinuationToken
			} else {
				return
			}
		}
	})
	return a.errSetup
}

func (a *auditHandler) getRepsitoryCreatedAt() (time.Time, error) {
	if err := a.setup(); err != nil {
		return time.Time{}, fmt.Errorf("error during auditHandler.setup: %w", err)
	}

	return a.createdAt, nil
}
