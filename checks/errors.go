// Copyright 2021 Security Scorecard Authors
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

package checks

import (
	"errors"
)

//nolint
var (
	errInternalInvalidDockerFile    = errors.New("invalid Dockerfile")
	errInternalInvalidYamlFile      = errors.New("invalid yaml file")
	errInternalFilenameMatch        = errors.New("filename match error")
	errInternalEmptyFile            = errors.New("empty file")
	errInvalidGitHubWorkflow        = errors.New("invalid GitHub workflow")
	errInternalNoReviews            = errors.New("no reviews found")
	errInternalNoCommits            = errors.New("no commits found")
	errInternalInvalidPermissions   = errors.New("invalid permissions")
	errInternalNameCannotBeEmpty    = errors.New("name cannot be empty")
	errInternalCheckFuncCannotBeNil = errors.New("checkFunc cannot be nil")
)
