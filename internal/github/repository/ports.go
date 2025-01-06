package repository

import (
	"context"

	"github.com/termkit/gama/internal/github/domain"
)

type Repository interface {
	ListRepositories(ctx context.Context, limit int, skip int, sort domain.SortBy) ([]GithubRepository, error)
	GetAuthUser(ctx context.Context) (*GithubUser, error)
	GetRepository(ctx context.Context, repository string) (*GithubRepository, error)
	ListBranches(ctx context.Context, repository string) ([]GithubBranch, error)
	ListWorkflowRuns(ctx context.Context, repository string) (*WorkflowRuns, error)
	TriggerWorkflow(ctx context.Context, repository string, branch string, workflowName string, workflow any) error
	GetWorkflows(ctx context.Context, repository string) ([]Workflow, error)
	GetTriggerableWorkflows(ctx context.Context, repository string) ([]Workflow, error)
	InspectWorkflowContent(ctx context.Context, repository string, branch string, workflowFile string) ([]byte, error)
	ReRunFailedJobs(ctx context.Context, repository string, runId int64) error
	ReRunWorkflow(ctx context.Context, repository string, runId int64) error
	CancelWorkflow(ctx context.Context, repository string, runId int64) error
}
