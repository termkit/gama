package usecase

import (
	"context"
)

type UseCase interface {
	ListRepositories(ctx context.Context, input ListRepositoriesInput) (*ListRepositoriesOutput, []error)
	ListWorkflowByRepository(ctx context.Context, input ListWorkflowByRepositoryInput) (*ListWorkflowByRepositoryOutput, error)
	ListWorkflowRuns(ctx context.Context, input ListWorkflowRunsInput) (*ListWorkflowRunsOutput, error)
	InspectWorkflow(ctx context.Context, input InspectWorkflowInput) (*InspectWorkflowOutput, error)
	TriggerWorkflow(ctx context.Context, input TriggerWorkflowInput) (*TriggerWorkflowOutput, error)
}
