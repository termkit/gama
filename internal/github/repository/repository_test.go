package repository

// TODO : Write mock tests for this package

import (
	"context"
	"os"
	"testing"
)

func newRepo(ctx context.Context) *Repo {
	repo := New()

	repo.Initialize(ctx, GithubConfig{Token: os.Getenv("GITHUB_TOKEN")})
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

	repositories, err := repo.ListRepositories(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(repositories) == 0 {
		t.Error("Expected repositories, got none")
	}
}

func TestRepo_ListBranches(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	branches, err := repo.ListBranches(ctx, "canack/railsgoat")
	if err != nil {
		t.Error(err)
	}

	t.Log(branches)
}

func TestRepo_ListWorkflowRuns(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	workflowRuns, err := repo.ListWorkflowRuns(ctx, "canack/tc", "master")
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
