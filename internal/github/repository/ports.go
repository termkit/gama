package repository

import (
	"context"
)

type Repository interface {
	Initialize(ctx context.Context, config GithubConfig, opts ...InitializeOptions)
	TestConnection(ctx context.Context) error
	ListRepositories(ctx context.Context) ([]GithubRepository, error)
	ListBranches(ctx context.Context, repository string) ([]GithubBranch, error)
	ListWorkflowRuns(ctx context.Context, repository string, branch string) (any, error)
	TriggerWorkflow(ctx context.Context, repository string, branch string, workflowName string, workflow any) error
	GetTriggerableWorkflows(ctx context.Context, repository string) ([]Workflow, error)
	InspectWorkflowContent(ctx context.Context, repository string, workflowFile string) ([]byte, error)
	GetWorkflowRun(ctx context.Context, repository string, runId int64) (GithubWorkflowRun, error)
	GetWorkflowRunLogs(ctx context.Context, repository string, runId int64) (GithubWorkflowRunLogs, error)
}
