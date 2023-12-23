package usecase

import (
	pw "github.com/termkit/gama/pkg/workflow"
)

type ListRepositoriesInput struct {
}

type ListRepositoriesOutput struct {
	Repositories []GithubRepository
}

type GithubRepository struct {
	Name          string
	Private       bool
	DefaultBranch string
	Stars         int

	TriggerableWorkflows []Workflow
	// We can add more fields here
}

// ------------------------------------------------------------

type ListWorkflowByRepositoryInput struct {
	Repository string
}

type ListWorkflowByRepositoryOutput struct {
	Workflows []Workflow
}
type Workflow struct {
	ID    int64
	Name  string
	State string
}

// ------------------------------------------------------------

type ListWorkflowRunsInput struct {
	Repository string
	Branch     string
}

type ListWorkflowRunsOutput struct {
	Workflows []Workflow
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
