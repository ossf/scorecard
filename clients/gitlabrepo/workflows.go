package gitlabrepo

import (
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type workflowsHandler struct {
	glClient *gitlab.Client
	repourl  *repoURL
}

func (handler *workflowsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
}

func (handler *workflowsHandler) listSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	jobs, _, err := handler.glClient.Jobs.ListProjectJobs(handler.repourl.projectID, &gitlab.ListJobsOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting project jobs: %w", err)
	}

	return workflowsRunsFrom(jobs, filename), nil
}

func workflowsRunsFrom(data []*gitlab.Job, filename string) []clients.WorkflowRun {
	var workflowRuns []clients.WorkflowRun
	for _, job := range data {
		// Find a better way to do this.
		for _, artifact := range job.Artifacts {
			if strings.EqualFold(artifact.Filename, filename) {
				workflowRuns = append(workflowRuns, clients.WorkflowRun{
					HeadSHA: &job.Pipeline.Sha,
					URL:     job.WebURL,
				})
				break
			}
		}
	}
	return workflowRuns
}
