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

	Workflows []Workflow
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

type InspectWorkflowInput struct {
	Repository   string
	Branch       string
	WorkflowFile string
}

type InspectWorkflowOutput struct {
	Workflow *pw.Pretty
}

// ------------------------------------------------------------

type TriggerWorkflowInput struct {
	WorkflowFile string
	Repository   string
	Branch       string
	Content      string // workflow content in json format
}

type TriggerWorkflowOutput struct {
	// Return workflow information
	// Like status url etc.
}

// ------------------------------------------------------------

type GetTriggerableWorkflowsInput struct {
	Repository string
	Branch     string
}

type GetTriggerableWorkflowsOutput struct {
	TriggerableWorkflows []TriggerableWorkflow
}

type TriggerableWorkflow struct {
	ID   int64
	Name string
	Path string
}

// ------------------------------------------------------------

type ReRunFailedJobsInput struct {
	Repository string
	WorkflowID int64
}

type ReRunFailedJobsOutput struct {
}

// ------------------------------------------------------------

type ReRunWorkflowInput struct {
	Repository string
	WorkflowID int64
}

type ReRunWorkflowOutput struct {
}

// ------------------------------------------------------------

type CancelWorkflowInput struct {
	Repository string
	WorkflowID int64
}

type CancelWorkflowOutput struct {
}
