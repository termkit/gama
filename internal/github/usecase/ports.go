package usecase

import (
	"context"
)

type UseCase interface {
	ListRepositories(ctx context.Context, input ListRepositoriesInput) (*ListRepositoriesOutput, error)
	GetWorkflowHistory(ctx context.Context, input GetWorkflowHistoryInput) (*GetWorkflowHistoryOutput, error)
	ListWorkflowRuns(ctx context.Context, input ListWorkflowRunsInput) (*ListWorkflowRunsOutput, error)
	InspectWorkflow(ctx context.Context, input InspectWorkflowInput) (*InspectWorkflowOutput, error)
	TriggerWorkflow(ctx context.Context, input TriggerWorkflowInput) (*TriggerWorkflowOutput, error)
}
