package handler

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/termkit/gama/internal/config"
	gu "github.com/termkit/gama/internal/github/usecase"
	pkgversion "github.com/termkit/gama/pkg/version"
	"github.com/termkit/skeleton"
)

func SetupTerminal(githubUseCase gu.UseCase, version pkgversion.Version) tea.Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	s := skeleton.NewSkeleton()

	s.AddPage("info", "Info", SetupModelInfo(s, githubUseCase, version))
	s.AddPage("repository", "Repository", SetupModelGithubRepository(s, githubUseCase))
	s.AddPage("history", "Workflow History", SetupModelGithubWorkflowHistory(s, githubUseCase))
	s.AddPage("workflow", "Workflow", SetupModelGithubWorkflow(s, githubUseCase))
	s.AddPage("trigger", "Trigger", SetupModelGithubTrigger(s, githubUseCase))

	s.SetBorderColor("#ff0055").
		SetActiveTabBorderColor("#ff0055").
		SetInactiveTabBorderColor("#82636f").
		SetWidgetBorderColor("#ff0055")

	if cfg.Settings.LiveMode.Enabled {
		s.AddWidget("live", "Live Mode: On")
	} else {
		s.AddWidget("live", "Live Mode: Off")
	}

	s.SetTerminalViewportWidth(MinTerminalWidth)
	s.SetTerminalViewportHeight(MinTerminalHeight)

	s.KeyMap.SetKeyNextTab(handlerKeys.SwitchTabRight)
	s.KeyMap.SetKeyPrevTab(handlerKeys.SwitchTabLeft)
	s.KeyMap.SetKeyQuit(handlerKeys.Quit)

	return s
}
