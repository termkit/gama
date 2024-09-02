package handler

import (
	tea "github.com/charmbracelet/bubbletea"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	pkgversion "github.com/termkit/gama/pkg/version"
	"github.com/termkit/skeleton"
	"time"
)

func SetupTerminal(githubUseCase gu.UseCase, version pkgversion.Version) tea.Model {
	s := skeleton.NewSkeleton()

	s.AddPage("info", "Info", SetupModelInfo(s, githubUseCase, version)).
		AddPage("repository", "Repository", SetupModelGithubRepository(s, githubUseCase)).
		AddPage("history", "Workflow History", SetupModelGithubWorkflowHistory(s, githubUseCase)).
		AddPage("workflow", "Workflow", SetupModelGithubWorkflow(s, githubUseCase)).
		AddPage("trigger", "Trigger", SetupModelGithubTrigger(s, githubUseCase))

	s.SetBorderColor("#ff0055").SetActiveTabBorderColor("#ff0055").SetInactiveTabBorderColor("#82636f").SetWidgetBorderColor("#ff0055")

	//s.AddWidget("version", "development mode")
	time.Sleep(100 * time.Millisecond)
	s.AddWidget("repositories", "Repository Count: 0")
	time.Sleep(100 * time.Millisecond)
	s.AddWidget("live", "Live Mode: Off")

	s.SetTerminalViewportWidth(hdltypes.MinTerminalWidth)
	s.SetTerminalViewportHeight(hdltypes.MinTerminalHeight)

	s.KeyMap.SetKeyNextTab(handlerKeys.SwitchTabRight)
	s.KeyMap.SetKeyPrevTab(handlerKeys.SwitchTabLeft)
	s.KeyMap.SetKeyQuit(handlerKeys.Quit)

	return s
}
