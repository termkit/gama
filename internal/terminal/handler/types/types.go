package types

type SelectedRepository struct {
	RepositoryID   int64  // repository id
	RepositoryName string // full repository name (owner/name)
	WorkflowName   string // workflow name
	BranchName     string // branch name
}

var ScreenWidth *int
