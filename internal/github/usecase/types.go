package usecase

import (
	"time"

	"github.com/termkit/gama/internal/github/domain"
	pw "github.com/termkit/gama/pkg/workflow"
)

type ListRepositoriesInput struct {
	Limit int
	Page  int
	Sort  domain.SortBy
}

func (i *ListRepositoriesInput) Prepare() {
	if i.Limit <= 0 {
		i.Limit = 500
	}

	if i.Page <= 0 {
		i.Page = 1
	}

	if i.Sort == "" {
		i.Sort = domain.SortByPushed
	}
}

type ListRepositoriesOutput struct {
	Repositories []GithubRepository
}

// ------------------------------------------------------------

type GetRepositoryBranchesInput struct {
	Repository string
}

type GetRepositoryBranchesOutput struct {
	Branches []GithubBranch
}

type GithubBranch struct {
	Name      string
	IsDefault bool
}

// ------------------------------------------------------------

type GithubRepository struct {
	Name          string
	Private       bool
	DefaultBranch string
	Stars         int
	LastUpdated   time.Time

	Workflows []Workflow
	// We can add more fields here
}

// ------------------------------------------------------------

type GetAuthUserOutput struct {
	GithubUser
}

type GithubUser struct {
	Login string `json:"login"` // username
	ID    int    `json:"id"`
	Email string `json:"email"`
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

// ------------------------------------------------------------

type ReRunWorkflowInput struct {
	Repository string
	WorkflowID int64
}

// ------------------------------------------------------------

type CancelWorkflowInput struct {
	Repository string
	WorkflowID int64
}
