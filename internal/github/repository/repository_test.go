package repository

// TODO : Write mock tests for this package

import (
	"context"
	"testing"

	pkgconfig "github.com/termkit/gama/pkg/config"
)

func newRepo(ctx context.Context) *Repo {
	cfg, err := pkgconfig.LoadConfig()
	if err != nil {
		panic(err)
	}

	repo := New(cfg)
	return repo
}

func TestRepo_TestConnection(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	err := repo.TestConnection(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestRepo_ListRepositories(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	repositories, err := repo.ListRepositories(ctx, 10)
	if err != nil {
		t.Error(err)
	}

	if len(repositories) == 0 {
		t.Error("Expected repositories, got none")
	}
}

func TestRepo_GetRepository(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	repository, err := repo.GetRepository(ctx, "canack/tc")
	if err != nil {
		t.Error(err)
	}

	t.Log(repository)
}

func TestRepo_ListBranches(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	branches, err := repo.ListBranches(ctx, "canack/tc")
	if err != nil {
		t.Error(err)
	}

	t.Log(branches)
}

func TestRepo_ListWorkflowRuns(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	targetRepositoryName := "canack/tc"
	targetRepository, err := repo.GetRepository(ctx, targetRepositoryName)
	if err != nil {
		t.Error(err)
	}

	defaultBranch := targetRepository.DefaultBranch

	workflowRuns, err := repo.ListWorkflowRuns(ctx, targetRepositoryName, defaultBranch)
	if err != nil {
		t.Error(err)
	}

	t.Log(workflowRuns)
}

func TestRepo_GetTriggerableWorkflows(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	workflows, err := repo.GetTriggerableWorkflows(ctx, "canack/tc")
	if err != nil {
		t.Error(err)
	}

	t.Log(workflows)
}
