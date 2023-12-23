package usecase

import (
	gt "github.com/termkit/gama/internal/github/types"
	pw "github.com/termkit/gama/pkg/workflow"
)

type ListRepositoriesInput struct {
}

type ListRepositoriesOutput struct {
	Repositories []gt.GithubRepository
}

// ------------------------------------------------------------

type ListWorkflowByRepositoryInput struct {
	Repository string
}

type ListWorkflowByRepositoryOutput struct {
	Workflows []gt.Workflow
}

// ------------------------------------------------------------

type ListWorkflowRunsInput struct {
	Repository string
	Branch     string
}

type ListWorkflowRunsOutput struct {
	Workflows []gt.Workflow
}

// ------------------------------------------------------------

type InspectWorkflowInput struct {
	Repository   string
	Branch       string
	WorkflowFile string
}

type InspectWorkflowOutput struct {
	Workflow pw.Workflow
}

// ------------------------------------------------------------

type TriggerWorkflowInput struct {
	Repository   string
	Branch       string
	WorkflowFile string
	Content      pw.Workflow
}

type TriggerWorkflowOutput struct {
	// Return workflow information
	// Like status url etc.
}
