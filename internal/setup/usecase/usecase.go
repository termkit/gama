package usecase

import (
	"context"
	"time"

	gr "github.com/termkit/gama/internal/github/repository"
	pkgconfig "github.com/termkit/gama/pkg/config"
)

type useCase struct {
	config           *pkgconfig.Config
	githubRepository gr.Repository
}

func New(githubRepository gr.Repository) UseCase {
	return &useCase{
		githubRepository: githubRepository,
	}
}

func (u *useCase) Setup(ctx context.Context) error {
	config, err := pkgconfig.LoadConfig()
	if err != nil {
		return err
	}

	u.config = config

	//if err := u.testConnection(ctx); err != nil {
	//	return err
	//}

	return nil
}

func (u *useCase) testConnection(ctx context.Context) error {
	var githubConfig = gr.GithubConfig{
		Token: u.config.Github.Token,
	}

	var initializeOptions = gr.InitializeOptions{
		HTTPTimeout: time.Second * 30,
	}

	u.githubRepository.Initialize(ctx, githubConfig, initializeOptions)

	return u.githubRepository.TestConnection(ctx)
}
