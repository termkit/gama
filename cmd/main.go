package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	gr "github.com/termkit/gama/internal/github/repository"
	gu "github.com/termkit/gama/internal/github/usecase"
	th "github.com/termkit/gama/internal/terminal/handler"
	vr "github.com/termkit/gama/internal/version/repository"
	vu "github.com/termkit/gama/internal/version/usecase"
	pkgconfig "github.com/termkit/gama/pkg/config"
)

var Version = "under development" // will be set by build flag

func main() {
	cfg, err := pkgconfig.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	githubRepository := gr.New(cfg)
	versionRepository := vr.New(Version)

	githubUseCase := gu.New(githubRepository)
	versionUseCase := vu.New(versionRepository)

	terminal := th.SetupTerminal(githubUseCase, versionUseCase)
	if _, err := tea.NewProgram(terminal).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
