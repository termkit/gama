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

func TestUseCase_InspectWorkflow(t *testing.T) {
	ctx := context.Background()

	githubRepo := repository.New()
	githubRepo.Initialize(ctx, repository.GithubConfig{Token: os.Getenv("GITHUB_TOKEN")})

	githubUseCase := New(githubRepo)

	workflow, err := githubUseCase.InspectWorkflow(ctx, InspectWorkflowInput{
		Repository:   "canack/tc",
		WorkflowFile: ".github/workflows/dispatch_test.yaml",
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(workflow)
}

func TestUseCase_TriggerWorkflow(t *testing.T) {
	ctx := context.Background()

	githubRepo := repository.New()
	githubRepo.Initialize(ctx, repository.GithubConfig{Token: os.Getenv("GITHUB_TOKEN")})

	githubUseCase := New(githubRepo)

	workflow, err := githubUseCase.InspectWorkflow(ctx, InspectWorkflowInput{
		Repository:   "canack/tc",
		WorkflowFile: ".github/workflows/dispatch_test.yaml",
	})

	for i, w := range workflow.Workflow.Inputs {
		if w.Key == "go-version" {
			w.SetValue("2.0")
			workflow.Workflow.Inputs[i] = w
		}
	}

	workflowJson, err := workflow.Workflow.ToJson()
	if err != nil {
		t.Error(err)
	}

	trigger, err := githubUseCase.TriggerWorkflow(ctx, TriggerWorkflowInput{
		WorkflowFile: ".github/workflows/dispatch_test.yaml",
		Repository:   "canack/tc",
		Branch:       "master",
		Content:      workflowJson,
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(trigger)
}
