package usecase

import (
	"context"
)

type UseCase interface {
	ListRepositories(ctx context.Context, input ListRepositoriesInput) (*ListRepositoriesOutput, error)
	GetWorkflowHistory(ctx context.Context, input GetWorkflowHistoryInput) (*GetWorkflowHistoryOutput, error)
	GetTriggerableWorkflows(ctx context.Context, input GetTriggerableWorkflowsInput) (*GetTriggerableWorkflowsOutput, error)
	InspectWorkflow(ctx context.Context, input InspectWorkflowInput) (*InspectWorkflowOutput, error)
	TriggerWorkflow(ctx context.Context, input TriggerWorkflowInput) (*TriggerWorkflowOutput, error)
}
