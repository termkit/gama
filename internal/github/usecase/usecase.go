package usecase

import (
	"context"
	"fmt"
	"time"

	gr "github.com/termkit/gama/internal/github/repository"
	pw "github.com/termkit/gama/pkg/workflow"
	py "github.com/termkit/gama/pkg/yaml"
)

type useCase struct {
	githubRepository gr.Repository
}

func New(githubRepository gr.Repository) UseCase {
	return &useCase{
		githubRepository: githubRepository,
	}
}

func (u useCase) ListRepositories(ctx context.Context, input ListRepositoriesInput) (*ListRepositoriesOutput, error) {
	repositories, err := u.githubRepository.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}

	// Create a buffered channel for jobs, results and errors
	jobs := make(chan gr.GithubRepository, len(repositories))
	results := make(chan GithubRepository, len(repositories))
	errors := make(chan error, len(repositories))

	// Start a number of workers
	for w := 1; w <= len(repositories); w++ {
		go u.workerListRepositories(ctx, jobs, results, errors)
	}

	// Send jobs to the workers
	for _, repository := range repositories {
		jobs <- repository
	}
	close(jobs)

	// Collect the results and errors
	var result []GithubRepository
	var resultErr error
	for range repositories {
		select {
		case res := <-results:
			result = append(result, res)
		case err := <-errors:
			resultErr = err
		}
	}

	return &ListRepositoriesOutput{
		Repositories: result,
	}, resultErr
}

func (u useCase) workerListRepositories(ctx context.Context, jobs <-chan gr.GithubRepository, results chan<- GithubRepository, errors chan<- error) {
	for repository := range jobs {
		getWorkflows, err := u.githubRepository.GetWorkflows(ctx, repository.FullName)
		if err != nil {
			errors <- err
			continue
		}

		var workflows []Workflow
		for _, workflow := range getWorkflows {
			workflows = append(workflows, Workflow{
				ID: workflow.ID,
			})
		}

		results <- GithubRepository{
			Name:          repository.FullName,
			Stars:         repository.StargazersCount,
			Private:       repository.Private,
			DefaultBranch: repository.DefaultBranch,
			Workflows:     workflows,
		}
	}
}

func (u useCase) GetWorkflowHistory(ctx context.Context, input GetWorkflowHistoryInput) (*GetWorkflowHistoryOutput, error) {
	var targetRepositoryName = input.Repository
	var targetBranch = input.Branch
	if targetBranch == "" {
		repository, err := u.githubRepository.GetRepository(ctx, targetRepositoryName)
		if err != nil {
			return nil, err
		}
		targetBranch = repository.DefaultBranch
	}

	workflowRuns, err := u.githubRepository.ListWorkflowRuns(ctx, targetRepositoryName, targetBranch)
	if err != nil {
		return nil, err
	}

	var workflows []Workflow
	for _, workflowRun := range workflowRuns.WorkflowRuns {
		workflows = append(workflows, Workflow{
			ID:           workflowRun.WorkflowID,
			WorkflowName: workflowRun.Name,
			ActionName:   workflowRun.DisplayTitle,
			TriggeredBy:  workflowRun.Actor.Login,
			StartedAt:    u.timeToString(workflowRun.CreatedAt),
			Status:       workflowRun.Status,
			Conclusion:   workflowRun.Conclusion,
			Duration:     u.getDuration(workflowRun.CreatedAt, workflowRun.UpdatedAt, workflowRun.Status),
		})
	}

	return &GetWorkflowHistoryOutput{
		Workflows: workflows,
	}, nil
}

func (u useCase) GetTriggerableWorkflows(ctx context.Context, input GetTriggerableWorkflowsInput) (*GetTriggerableWorkflowsOutput, error) {
	// TODO: Add branch option
	triggerableWorkflows, err := u.githubRepository.GetTriggerableWorkflows(ctx, input.Repository)
	if err != nil {
		return nil, err
	}

	var workflows []TriggerableWorkflow
	for _, workflow := range triggerableWorkflows {
		workflows = append(workflows, TriggerableWorkflow{
			ID:   workflow.ID,
			Name: workflow.Name,
			Path: workflow.Path,
		})
	}

	return &GetTriggerableWorkflowsOutput{
		TriggerableWorkflows: workflows,
	}, nil
}

func (u useCase) InspectWorkflow(ctx context.Context, input InspectWorkflowInput) (*InspectWorkflowOutput, error) {
	workflowData, err := u.githubRepository.InspectWorkflowContent(ctx, input.Repository, input.WorkflowFile)
	if err != nil {
		return nil, err
	}

	workflowContent, err := py.UnmarshalWorkflowContent(workflowData)
	if err != nil {
		return nil, err
	}

	workflow, err := pw.ParseWorkflow(*workflowContent)
	if err != nil {
		return nil, err
	}

	pretty := workflow.ToPretty()

	return &InspectWorkflowOutput{
		Workflow: pretty,
	}, nil
}

func (u useCase) TriggerWorkflow(ctx context.Context, input TriggerWorkflowInput) (*TriggerWorkflowOutput, error) {
	err := u.githubRepository.TriggerWorkflow(ctx, input.Repository, input.Branch, input.WorkflowFile, input.Content)
	if err != nil {
		return nil, err
	}

	return &TriggerWorkflowOutput{}, nil
}

func (u useCase) timeToString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func (u useCase) getDuration(startTime time.Time, endTime time.Time, status string) string {
	if status != "completed" {
		return "running"
	}

	diff := endTime.Sub(startTime)

	if diff.Seconds() < 60 {
		return fmt.Sprintf("%ds", int(diff.Seconds()))
	} else if diff.Seconds() < 3600 {
		return fmt.Sprintf("%dm %ds", int(diff.Minutes()), int(diff.Seconds())%60)
	} else {
		return fmt.Sprintf("%dh %dm %ds", int(diff.Hours()), int(diff.Minutes())%60, int(diff.Seconds())%60)
	}
}
