package handler

import (
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlgithubrepo "github.com/termkit/gama/internal/terminal/handler/ghrepository"
	hdltrigger "github.com/termkit/gama/internal/terminal/handler/ghtrigger"
	hdlWorkflow "github.com/termkit/gama/internal/terminal/handler/ghworkflow"
	hdlworkflowhistory "github.com/termkit/gama/internal/terminal/handler/ghworkflowhistory"
	hdlinfo "github.com/termkit/gama/internal/terminal/handler/information"
	hdlskeleton "github.com/termkit/gama/internal/terminal/handler/skeleton"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	ts "github.com/termkit/gama/internal/terminal/style"
	pkgversion "github.com/termkit/gama/pkg/version"
)

type model struct {
	// models
	viewport      *viewport.Model
	modelSkeleton *hdlskeleton.Skeleton

	// keymap
	keys keyMap
}

func SetupTerminal(githubUseCase gu.UseCase, version pkgversion.Version) tea.Model {
	skeleton := hdlskeleton.NewSkeleton()

	skeleton.AddPage(hdlskeleton.Title{Title: "Info", Style: hdlskeleton.TitleStyle{
		Active:   ts.TitleStyleActive,
		Inactive: ts.TitleStyleInactive,
	}}, hdlinfo.SetupModelInfo(githubUseCase, version))

	skeleton.AddPage(hdlskeleton.Title{Title: "Repository", Style: hdlskeleton.TitleStyle{
		Active:   ts.TitleStyleActive,
		Inactive: ts.TitleStyleInactive,
	}}, hdlgithubrepo.SetupModelGithubRepository(githubUseCase))

	skeleton.AddPage(hdlskeleton.Title{Title: "Workflow History", Style: hdlskeleton.TitleStyle{
		Active:   ts.TitleStyleActive,
		Inactive: ts.TitleStyleInactive,
	}}, hdlworkflowhistory.SetupModelGithubWorkflowHistory(githubUseCase))

	skeleton.AddPage(hdlskeleton.Title{Title: "Workflow", Style: hdlskeleton.TitleStyle{
		Active:   ts.TitleStyleActive,
		Inactive: ts.TitleStyleInactive,
	}}, hdlWorkflow.SetupModelGithubWorkflow(githubUseCase))

	skeleton.AddPage(hdlskeleton.Title{Title: "Trigger", Style: hdlskeleton.TitleStyle{
		Active:   ts.TitleStyleActive,
		Inactive: ts.TitleStyleInactive,
	}}, hdltrigger.SetupModelGithubTrigger(githubUseCase))

	m := model{
		viewport:      hdltypes.NewTerminalViewport(),
		modelSkeleton: skeleton,
		keys:          keys,
	}

	return &m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.SetWindowTitle("GitHub Actions Manager (GAMA)"),
		m.modelSkeleton.Init(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.modelSkeleton, cmd = m.modelSkeleton.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.viewport.Width < hdltypes.MinTerminalWidth || m.viewport.Height < hdltypes.MinTerminalHeight {
		return fmt.Sprintf("Terminal window is too small. Please resize to at least %dx%d.", hdltypes.MinTerminalWidth, hdltypes.MinTerminalHeight)
	}

	return m.modelSkeleton.View()
}
