package handler

import (
	tea "github.com/charmbracelet/bubbletea"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlgithubrepo "github.com/termkit/gama/internal/terminal/handler/ghrepository"
	hdltrigger "github.com/termkit/gama/internal/terminal/handler/ghtrigger"
	hdlWorkflow "github.com/termkit/gama/internal/terminal/handler/ghworkflow"
	hdlworkflowhistory "github.com/termkit/gama/internal/terminal/handler/ghworkflowhistory"
	hdlinfo "github.com/termkit/gama/internal/terminal/handler/information"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	pkgversion "github.com/termkit/gama/pkg/version"
	"github.com/termkit/skeleton"
)

func SetupTerminal(githubUseCase gu.UseCase, version pkgversion.Version) tea.Model {
	s := skeleton.NewSkeleton()

	s.AddPage("info", "Info", hdlinfo.SetupModelInfo(s, githubUseCase, version)).
		AddPage("repository", "Repository", hdlgithubrepo.SetupModelGithubRepository(s, githubUseCase)).
		AddPage("history", "Workflow History", hdlworkflowhistory.SetupModelGithubWorkflowHistory(s, githubUseCase)).
		AddPage("workflow", "Workflow", hdlWorkflow.SetupModelGithubWorkflow(s, githubUseCase)).
		AddPage("trigger", "Trigger", hdltrigger.SetupModelGithubTrigger(s, githubUseCase))

	s.SetBorderColor("49").SetActiveTabBorderColor("#ff0055")

	s.AddWidget("version", "development mode")

	s.SetTerminalViewportWidth(hdltypes.MinTerminalWidth)
	s.SetTerminalViewportHeight(hdltypes.MinTerminalHeight)

	s.KeyMap.SetKeyNextTab(keys.SwitchTabRight)
	s.KeyMap.SetKeyPrevTab(keys.SwitchTabLeft)
	s.KeyMap.SetKeyQuit(keys.Quit)

	return s
}
