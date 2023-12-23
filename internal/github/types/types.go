package types

type GithubRepository struct {
	Name          string
	Private       bool
	DefaultBranch string

	TriggerableWorkflows []Workflow

	// We can add more fields here
}

type Workflow struct {
	ID    int64
	Name  string
	State string
}
