package usecase

import (
	"context"
	"os"
	"testing"

	"github.com/termkit/gama/internal/github/repository"
)

func TestUseCase_ListRepositories(t *testing.T) {
	ctx := context.Background()

	githubRepo := repository.New()
	githubRepo.Initialize(ctx, repository.GithubConfig{Token: os.Getenv("GITHUB_TOKEN")})

	githubUseCase := New(githubRepo)

	repositories, err := githubUseCase.ListRepositories(ctx, ListRepositoriesInput{})
	if err != nil {
		t.Error(err)
	}
	t.Log(repositories)
}
