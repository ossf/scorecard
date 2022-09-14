package gitlabrepo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

var (
	_                     clients.RepoClient = &Client{}
	errInputRepoType                         = errors.New("input repo should be of type repoURL")
	errDefaultBranchEmpty                    = errors.New("default branch name is empty")
)

type Client struct {
	repourl       *repoURL
	repo          *gitlab.Project
	glClient      *gitlab.Client
	contributors  *contributorsHandler
	branches      *branchesHandler
	releases      *releasesHandler
	workflows     *workflowsHandler
	checkruns     *checkrunsHandler
	commits       *commitsHandler
	issues        *issuesHandler
	project       *projectHandler
	statuses      *statusesHandler
	search        *searchHandler
	searchCommits *searchCommitsHandler
	webhook       *webhookHandler
	languages     *languagesHandler
	ctx           context.Context
	tarball       tarballHandler
}

// InitRepo sets up the GitLab project in local storage for improving performance and GitLab token usage efficiency.
func (client *Client) InitRepo(inputRepo clients.Repo, commitSHA string) error {
	glRepo, ok := inputRepo.(*repoURL)
	if !ok {
		return fmt.Errorf("%w: %v", errInputRepoType, inputRepo)
	}

	// Sanity check.
	repo, _, err := client.glClient.Projects.GetProject(glRepo.projectID, &gitlab.GetProjectOptions{})
	if err != nil {
		return sce.WithMessage(sce.ErrRepoUnreachable, err.Error())
	}

	client.repo = repo
	client.repourl = &repoURL{
		hostname:      inputRepo.URI(),
		owner:         repo.Owner.Username,
		projectID:     fmt.Sprint(repo.ID),
		defaultBranch: repo.DefaultBranch,
		commitSHA:     commitSHA,
	}

	// Init contributorsHandler
	client.contributors.init(client.repourl)

	// Init commitsHandler
	client.commits.init(client.repourl)

	// Init branchesHandler
	client.branches.init(client.repourl)

	// Init releasesHandler
	client.releases.init(client.repourl)

	// Init issuesHandler
	client.issues.init(client.repourl)

	// Init projectHandler
	client.project.init(client.repourl)

	// Init workflowsHandler
	client.workflows.init(client.repourl)

	// Init checkrunsHandler
	client.checkruns.init(client.repourl)

	// Init statusesHandler
	client.statuses.init(client.repourl)

	// Init searchHandler
	client.search.init(client.repourl)

	// Init searchCommitsHandler
	client.searchCommits.init(client.repourl)

	// Init webhookHandler
	client.webhook.init(client.repourl)

	// Init languagesHandler
	client.languages.init(client.repourl)

	// Init tarballHandler.
	client.tarball.init(client.ctx, client.repourl, client.repo, commitSHA)
	return nil
}

func (client *Client) URI() string {
	return fmt.Sprintf("%s/%s/%s", client.repourl.hostname, client.repourl.owner, client.repourl.projectID)
}

func (client *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return nil, nil
	// return client.tarball.listFiles(predicate)
}

func (client *Client) GetFileContent(filename string) ([]byte, error) {
	return nil, nil
	// return client.tarball.getFileContent(filename)
}

func (client *Client) ListCommits() ([]clients.Commit, error) {
	return client.commits.listCommits()
}

func (client *Client) ListIssues() ([]clients.Issue, error) {
	return client.issues.listIssues()
}

func (client *Client) ListReleases() ([]clients.Release, error) {
	return client.releases.getReleases()
}

func (client *Client) ListContributors() ([]clients.User, error) {
	return client.contributors.getContributors()
}

func (client *Client) IsArchived() (bool, error) {
	return client.project.isArchived()
}

func (client *Client) GetDefaultBranch() (*clients.BranchRef, error) {
	branch, err := client.branches.getDefaultBranch()
	if err != nil {
		return nil, fmt.Errorf("%w", errDefaultBranchEmpty)
	}
	return branch, nil
}

func (client *Client) GetDefaultBranchName() (string, error) {
	return client.repourl.defaultBranch, nil
}

func (client *Client) GetBranch(branch string) (*clients.BranchRef, error) {
	return client.branches.getBranch(branch)
}

func (client *Client) GetCreatedAt() (time.Time, error) {
	return client.project.getCreatedAt()
}

func (client *Client) ListWebhooks() ([]clients.Webhook, error) {
	return client.webhook.listWebhooks()
}

func (client *Client) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return client.workflows.listSuccessfulWorkflowRuns(filename)
}

func (client *Client) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return client.checkruns.listCheckRunsForRef(ref)
}

func (client *Client) ListStatuses(ref string) ([]clients.Status, error) {
	return client.statuses.listStatuses(ref)
}

func (client *Client) ListProgrammingLanguages() ([]clients.Language, error) {
	return client.languages.listProgrammingLanguages()
}

func (client *Client) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return client.search.search(request)
}

func (client *Client) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return client.searchCommits.search(request)
}

func (client *Client) Close() error {
	return client.tarball.cleanup()
}

func CreateGitlabClientWithToken(ctx context.Context, token string, repo clients.Repo) (clients.RepoClient, error) {
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(repo.URI()))
	if err != nil {
		return nil, fmt.Errorf("could not create gitlab client with error: %w", err)
	}

	return &Client{
		ctx:      ctx,
		glClient: client,
		contributors: &contributorsHandler{
			glClient: client,
		},
		branches: &branchesHandler{
			glClient: client,
		},
		releases: &releasesHandler{
			glClient: client,
		},
		workflows: &workflowsHandler{
			glClient: client,
		},
		checkruns: &checkrunsHandler{
			glClient: client,
		},
		commits: &commitsHandler{
			glClient: client,
		},
		issues: &issuesHandler{
			glClient: client,
		},
		project: &projectHandler{
			glClient: client,
		},
		statuses: &statusesHandler{
			glClient: client,
		},
		search: &searchHandler{
			glClient: client,
		},
		searchCommits: &searchCommitsHandler{
			glClient: client,
		},
		webhook: &webhookHandler{
			glClient: client,
		},
		languages: &languagesHandler{
			glClient: client,
		},
	}, nil
}

// TODO: implement CreateOssFuzzRepoClient.
func CreateOssFuzzRepoClient(ctx context.Context, logger *log.Logger) (clients.RepoClient, error) {
	return nil, fmt.Errorf("%w, oss fuzz currently only supported for github repos", clients.ErrUnsupportedFeature)
}
