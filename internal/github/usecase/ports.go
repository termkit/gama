package usecase

import (
	"context"

	"github.com/termkit/gama/pkg/pagination"
)

type UseCase interface {
	ListRepositories(ctx context.Context, input pagination.FindOpts) (*ListRepositoriesOutput, error)
	GetWorkflowHistory(ctx context.Context, input GetWorkflowHistoryInput) (*GetWorkflowHistoryOutput, error)
	GetTriggerableWorkflows(ctx context.Context, input GetTriggerableWorkflowsInput) (*GetTriggerableWorkflowsOutput, error)
	InspectWorkflow(ctx context.Context, input InspectWorkflowInput) (*InspectWorkflowOutput, error)
	TriggerWorkflow(ctx context.Context, input TriggerWorkflowInput) (*TriggerWorkflowOutput, error)
	ReRunFailedJobs(ctx context.Context, input ReRunFailedJobsInput) (*ReRunFailedJobsOutput, error)
	ReRunWorkflow(ctx context.Context, input ReRunWorkflowInput) (*ReRunWorkflowOutput, error)
	CancelWorkflow(ctx context.Context, input CancelWorkflowInput) (*CancelWorkflowOutput, error)
}
