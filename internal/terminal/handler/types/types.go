package types

type SelectedRepository struct {
	RepositoryName string // full repository name (owner/name)
	WorkflowName   string // workflow name
	BranchName     string // branch name
}

var ScreenWidth *int
