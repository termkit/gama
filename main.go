package main

import (
	"fmt"
	"os"

	"github.com/termkit/gama/internal/config"
	gr "github.com/termkit/gama/internal/github/repository"
	gu "github.com/termkit/gama/internal/github/usecase"
	th "github.com/termkit/gama/internal/terminal/handler"
	pkgversion "github.com/termkit/gama/pkg/version"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	repositoryOwner = "termkit"
	repositoryName  = "gama"
)

var Version = "under development" // will be set by build flag

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	version := pkgversion.New(repositoryOwner, repositoryName, Version)

	githubRepository := gr.New(cfg)
	githubUseCase := gu.New(githubRepository)

	terminal := th.SetupTerminal(githubUseCase, version)
	if _, err := tea.NewProgram(terminal).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
