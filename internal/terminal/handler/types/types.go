package types

import (
	"sync"
)

type SelectedRepository struct {
	RepositoryName string // full repository name (owner/name)
	WorkflowName   string // workflow name
	BranchName     string // branch name
}

var (
	onceSelectedRepository sync.Once
	selectedRepository     *SelectedRepository
)

func NewSelectedRepository() *SelectedRepository {
	onceSelectedRepository.Do(func() {
		selectedRepository = &SelectedRepository{}
	})
	return selectedRepository
}

// ----------------------------------------------

const (
	MinTerminalWidth  = 102
	MinTerminalHeight = 24
)
