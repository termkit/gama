package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	gr "github.com/termkit/gama/internal/github/repository"
	gu "github.com/termkit/gama/internal/github/usecase"
	su "github.com/termkit/gama/internal/setup/usecase"
	th "github.com/termkit/gama/internal/terminal/handler"
)

func main() {
	var ctx = context.Background()

	githubRepository := gr.New()
	githubRepository.Initialize(ctx, gr.GithubConfig{Token: os.Getenv("GITHUB_TOKEN")})

	setupUseCase := su.New(githubRepository)

	if err := setupUseCase.Setup(ctx); err != nil {
		panic(fmt.Sprintf("failed to setup gama: %v", err))
	}

	githubUseCase := gu.New(githubRepository)

	terminal := th.SetupTerminal(githubUseCase)
	if _, err := tea.NewProgram(terminal).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
