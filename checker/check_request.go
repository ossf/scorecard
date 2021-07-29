// Copyright 2020 Security Scorecard Authors
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

package checker

import (
	"context"
	"net/http"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v2/clients"
)

// CheckRequest struct encapsulates all data to be passed into a CheckFn.
type CheckRequest struct {
	Ctx         context.Context
	Client      *github.Client
	GraphClient *githubv4.Client
	HTTPClient  *http.Client
	RepoClient  clients.RepoClient
	Dlogger     DetailLogger
	Owner, Repo string
}
