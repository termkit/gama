package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	gr "github.com/termkit/gama/internal/github/repository"
	gu "github.com/termkit/gama/internal/github/usecase"
	th "github.com/termkit/gama/internal/terminal/handler"
	pkgconfig "github.com/termkit/gama/pkg/config"
)

func main() {
	cfg, err := pkgconfig.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	githubRepository := gr.New(cfg)

	githubUseCase := gu.New(githubRepository)

	terminal := th.SetupTerminal(githubUseCase)
	if _, err := tea.NewProgram(terminal).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
