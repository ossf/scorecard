// NOTE: In Gitlab repositories are called projects, however to ensure compatibility,
// this package will regard to Gitlab projects as repositories.
package gitlabrepo

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	gitlabOrgProj = ".gitlab"
)

// TODO: check that this repoURL conforms to gitlab naming.
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
	*  "<companyDomain:string>/<owner:string>/<projectID:int>"
	*  "gitlab.<companyDomain:string>.com/<owner:string>/<projectID:int>"

* TODO: add support for following input string formats
	*  "<companyDomain:string>/<owner:string>/<URLEncodedPath:string>"
	*  "gitlab.<companyDomain:string>.com/<owner:string>/<URLEncodedPath:string>"
*/
func (r *repoURL) parse(input string) error {
	var t string

	const three = 3
	const four = 4

	c := strings.Split(input, "/")

	switch l := len(c); {
	// Sanitising the inputs to always be of case 3 or 4.
	case l == three:
		t = c[0] + "/" + c[1] + "/" + c[2]
	case l == four:
		t = input
	}

	if !strings.Contains(t, "://") {
		t = "https://" + t
	}

	u, e := url.Parse(t)
	if e != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", e))
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(u.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		return sce.WithMessage(sce.ErrorInvalidURL, fmt.Sprintf("%v. Expected full repository url", input))
	}

	// TODO: make sure this actually gets the the correct information.
	r.hostname, r.owner, r.projectID = u.Host, split[0], split[1]
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

// TODO: figure out what this function does and whether or not it is entirely necessary.
func (r *repoURL) Org() clients.Repo {
	return &repoURL{
		hostname:  r.hostname,
		owner:     r.owner,
		projectID: gitlabOrgProj,
	}
}

// IsValid implements Repo.IsValid.
func (r *repoURL) IsValid() error {
	if !strings.Contains(r.hostname, "gitlab.") {
		return sce.WithMessage(sce.ErrorUnsupportedHost, r.hostname)
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
