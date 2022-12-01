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

// NOTE: In Gitlab repositories are called projects, however to ensure compatibility,
// this package will regard to Gitlab projects as repositories.
package gitlabrepo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	gitlabOrgProj = ".gitlab"
)

type repoURL struct {
	hostname      string
	owner         string
	projectID     string
	defaultBranch string
	commitSHA     string
	metadata      []string
}

// Parses input string into repoURL struct
/*
*  Accepted input string formats are as follows:
	*  "gitlab.<companyDomain:string>.com/<owner:string>/<projectID:int>"
	* "https://gitlab.<companyDomain:string>.com/<owner:string>/<projectID:int>"
*/
func (r *repoURL) parse(input string) error {
	switch {
	case strings.Contains(input, "https://"):
		input = strings.TrimPrefix(input, "https://")
	case strings.Contains(input, "http://"):
		input = strings.TrimPrefix(input, "http://")
	case strings.Contains(input, "://"):
		return sce.WithMessage(sce.ErrScorecardInternal, "unknown input format")
	}

	stringParts := strings.Split(input, "/")

	stringParts[2] = strings.TrimSuffix(stringParts[2], "/")

	r.hostname, r.owner, r.projectID = stringParts[0], stringParts[1], stringParts[2]
	return nil
}

// URI implements Repo.URI().
// TODO: there may be a reason the string was originally in format "%s/%s/%s", hostname, owner, projectID,
// however I changed it to be more "userful".
func (r *repoURL) URI() string {
	return fmt.Sprintf("https://%s", r.hostname)
}

// String implements Repo.String.
func (r *repoURL) String() string {
	return fmt.Sprintf("%s-%s_%s", r.hostname, r.owner, r.projectID)
}

func (r *repoURL) Org() clients.Repo {
	return &repoURL{
		hostname:  r.hostname,
		owner:     r.owner,
		projectID: gitlabOrgProj,
	}
}

// IsValid implements Repo.IsValid.
func (r *repoURL) IsValid() error {
	hostMatched, err := regexp.MatchString("gitlab.*com", r.hostname)
	if err != nil {
		return fmt.Errorf("error processing regex: %w", err)
	}
	if !hostMatched {
		return sce.WithMessage(sce.ErrorInvalidURL, "non gitlab repository found")
	}

	isNotDigit := func(c rune) bool { return c < '0' || c > '9' }
	b := strings.IndexFunc(r.projectID, isNotDigit) == -1
	if !b {
		return sce.WithMessage(sce.ErrorInvalidURL, "incorrect format for projectID")
	}

	if strings.TrimSpace(r.owner) == "" || strings.TrimSpace(r.projectID) == "" {
		return sce.WithMessage(sce.ErrorInvalidURL,
			fmt.Sprintf("%v. Expected the full project url", r.URI()))
	}
	return nil
}

func (r *repoURL) AppendMetadata(metadata ...string) {
	r.metadata = append(r.metadata, metadata...)
}

// Metadata implements Repo.Metadata.
func (r *repoURL) Metadata() []string {
	return r.metadata
}

// MakeGitlabRepo takes input of forms in parse and returns and implementation
// of clients.Repo interface.
func MakeGitlabRepo(input string) (clients.Repo, error) {
	var repo repoURL
	if err := repo.parse(input); err != nil {
		return nil, fmt.Errorf("error during parse: %w", err)
	}
	if err := repo.IsValid(); err != nil {
		return nil, fmt.Errorf("error n IsValid: %w", err)
	}
	return &repo, nil
}
