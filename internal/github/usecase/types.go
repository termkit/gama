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

type GetWorkflowHistoryInput struct {
	Repository string
	Branch     string
}

type GetWorkflowHistoryOutput struct {
	Workflows []Workflow
}

type Workflow struct {
	ID           int64  // workflow id
	WorkflowName string // workflow name
	ActionName   string // commit message
	TriggeredBy  string // who triggered this workflow
	StartedAt    string // workflow's started at
	Status       string // workflow's status, like success, failure, etc.
	Conclusion   string // workflow's conclusion, like success, failure, etc.
	Duration     string // workflow's duration
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
