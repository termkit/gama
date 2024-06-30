package types

import (
	"github.com/charmbracelet/bubbles/viewport"

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

var ScreenWidth *int

// ----------------------------------------------

const (
	MinTerminalWidth  = 102
	MinTerminalHeight = 24
)

var (
	onceViewport sync.Once
	vp           *viewport.Model
)

func NewTerminalViewport() *viewport.Model {
	onceViewport.Do(func() {
		vp = &viewport.Model{Width: MinTerminalWidth, Height: MinTerminalHeight}
	})
	return vp
}
