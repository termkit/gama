package usecase

import (
	"context"
	"testing"

	"github.com/termkit/gama/internal/config"

	"github.com/termkit/gama/internal/github/domain"
	"github.com/termkit/gama/internal/github/repository"
)

func TestUseCase_ListRepositories(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	githubRepo := repository.New(cfg)

	githubUseCase := New(githubRepo)

	repositories, err := githubUseCase.ListRepositories(ctx, ListRepositoriesInput{
		Limit: 10,
		Page:  0,
		Sort:  domain.SortByCreated,
	})

	if err != nil {
		t.Error(err)
	}
	t.Log(repositories)
}

func TestUseCase_InspectWorkflow(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	githubRepo := repository.New(cfg)

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
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	githubRepo := repository.New(cfg)

	githubUseCase := New(githubRepo)

	workflow, err := githubUseCase.InspectWorkflow(ctx, InspectWorkflowInput{
		Repository:   "canack/tc",
		WorkflowFile: ".github/workflows/dispatch_test.yaml",
	})
	if err != nil {
		t.Error(err)
	}

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
