package usecase

import (
	"context"
)

type UseCase interface {
	GetAuthUser(ctx context.Context) (*GetAuthUserOutput, error)
	ListRepositories(ctx context.Context, input ListRepositoriesInput) (*ListRepositoriesOutput, error)
	GetWorkflowHistory(ctx context.Context, input GetWorkflowHistoryInput) (*GetWorkflowHistoryOutput, error)
	GetTriggerableWorkflows(ctx context.Context, input GetTriggerableWorkflowsInput) (*GetTriggerableWorkflowsOutput, error)
	InspectWorkflow(ctx context.Context, input InspectWorkflowInput) (*InspectWorkflowOutput, error)
	TriggerWorkflow(ctx context.Context, input TriggerWorkflowInput) (*TriggerWorkflowOutput, error)
	ReRunFailedJobs(ctx context.Context, input ReRunFailedJobsInput) (*ReRunFailedJobsOutput, error)
	ReRunWorkflow(ctx context.Context, input ReRunWorkflowInput) (*ReRunWorkflowOutput, error)
	CancelWorkflow(ctx context.Context, input CancelWorkflowInput) (*CancelWorkflowOutput, error)
}
