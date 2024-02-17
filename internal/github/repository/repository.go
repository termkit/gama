package repository

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/termkit/gama/internal/github/domain"
	pkgconfig "github.com/termkit/gama/pkg/config"

	"gopkg.in/yaml.v3"
)

type Repo struct {
	Client HttpClient

	githubToken string
}

var githubAPIURL = "https://api.github.com"

func New(cfg *pkgconfig.Config) *Repo {
	return &Repo{
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
		githubToken: cfg.Github.Token,
	}
}

func (r *Repo) TestConnection(ctx context.Context) error {
	// List repositories for the authenticated user
	var repositories []GithubRepository
	err := r.do(ctx, nil, &repositories, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/user/repos",
		contentType: "application/json",
		queryParams: map[string]string{
			"visibility": "all",
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) ListRepositories(ctx context.Context, limit int, page int, sort domain.SortBy) ([]GithubRepository, error) {
	resultsChan := make(chan []GithubRepository)
	errChan := make(chan error)

	for page := 1; page <= 5; page++ {
		go r.workerListRepositories(ctx, limit, page, sort, resultsChan, errChan)
	}

	var repositories []GithubRepository
	var repoErr error

	for range make([]int, 5) {
		select {
		case err := <-errChan:
			repoErr = errors.Join(err)
		case res := <-resultsChan:
			repositories = append(repositories, res...)
		}
	}

	if repoErr != nil {
		return nil, repoErr
	}

	return repositories, nil
}

func (r *Repo) workerListRepositories(ctx context.Context, limit int, page int, sort domain.SortBy, results chan<- []GithubRepository, errs chan<- error) {
	var repositories []GithubRepository
	err := r.do(ctx, nil, &repositories, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/user/repos",
		contentType: "application/json",
		queryParams: map[string]string{
			"visibility": "all",
			"per_page":   strconv.Itoa(limit),
			"page":       strconv.Itoa(page),
			"sort":       sort.String(),
			"direction":  "desc",
		},
	})
	if err != nil {
		errs <- err
		return
	}

	results <- repositories
}

func (r *Repo) ListBranches(ctx context.Context, repository string) ([]GithubBranch, error) {
	// List branches for the given repository
	var branches any
	err := r.do(ctx, nil, &branches, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository + "/branches",
		contentType: "application/json",
	})
	if err != nil {
		return nil, err
	}

	return []GithubBranch{}, nil
}

func (r *Repo) GetRepository(ctx context.Context, repository string) (*GithubRepository, error) {
	var repo GithubRepository
	err := r.do(ctx, nil, &repo, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository,
		contentType: "application/json",
	})
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func (r *Repo) ListWorkflowRuns(ctx context.Context, repository string, branch string) (*WorkflowRuns, error) {
	// List workflow runs for the given repository and branch
	var workflowRuns WorkflowRuns
	err := r.do(ctx, nil, &workflowRuns, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository + "/actions/runs",
		contentType: "application/json",
		queryParams: map[string]string{
			"branch": branch,
		},
	})
	if err != nil {
		return nil, err
	}

	return &workflowRuns, nil
}

func (r *Repo) TriggerWorkflow(ctx context.Context, repository string, branch string, workflowName string, workflow any) error {
	var payload = fmt.Sprintf(`{"ref": "%s", "inputs": %s}`, branch, workflow)

	// Trigger a workflow for the given repository and branch
	err := r.do(ctx, payload, nil, requestOptions{
		method: http.MethodPost,
		path:   githubAPIURL + "/repos/" + repository + "/actions/workflows/" + path.Base(workflowName) + "/dispatches",
		accept: "application/vnd.github+json",
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) GetWorkflows(ctx context.Context, repository string) ([]Workflow, error) {
	// Get a workflow run for the given repository and runId
	var githubWorkflow githubWorkflow
	err := r.do(ctx, nil, &githubWorkflow, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository + "/actions/workflows",
		contentType: "application/json",
	})
	if err != nil {
		return nil, err
	}

	return githubWorkflow.Workflows, nil
}

func (r *Repo) GetTriggerableWorkflows(ctx context.Context, repository string) ([]Workflow, error) {
	// Get a workflow run for the given repository and runId
	var workflows githubWorkflow
	err := r.do(ctx, nil, &workflows, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository + "/actions/workflows",
		contentType: "application/json",
	})
	if err != nil {
		return nil, err
	}

	// Create a buffered channel for results and errors
	results := make(chan *Workflow, len(workflows.Workflows))
	errs := make(chan error, len(workflows.Workflows))

	// Filter workflows to only include those that are dispatchable and manually triggerable
	for _, workflow := range workflows.Workflows {
		go r.workerGetTriggerableWorkflows(ctx, repository, workflow, results, errs)
	}

	// Collect the results and errors
	var result []Workflow
	var resultErrs []error
	for range workflows.Workflows {
		select {
		case res := <-results:
			// append only triggerable (dispatch) workflows
			if res != nil {
				result = append(result, *res)
			}
		case err := <-errs:
			resultErrs = append(resultErrs, err)
		}
	}

	return result, errors.Join(resultErrs...)
}

func (r *Repo) workerGetTriggerableWorkflows(ctx context.Context, repository string, workflow Workflow, results chan<- *Workflow, errs chan<- error) {
	// Get the workflow file content
	fileContent, err := r.getWorkflowFile(ctx, repository, workflow.Path)
	if err != nil {
		errs <- err
		return
	}

	// Parse the workflow file content as YAML
	var wfFile workflowFile
	err = yaml.Unmarshal([]byte(fileContent), &wfFile)
	if err != nil {
		errs <- err
		return
	}

	var dispatchWorkflow *Workflow

	// Check if the workflow file content has a "workflow_dispatch" key
	if _, ok := wfFile.On["workflow_dispatch"]; ok {
		dispatchWorkflow = &workflow
	}

	results <- dispatchWorkflow
}

func (r *Repo) InspectWorkflowContent(ctx context.Context, repository string, branch string, workflowFile string) ([]byte, error) {
	// Get the content of the workflow file
	var githubFile githubFile
	err := r.do(ctx, nil, &githubFile, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository + "/contents/" + workflowFile,
		contentType: "application/vnd.github.VERSION.raw",
		queryParams: map[string]string{
			"ref": branch,
		},
	})
	if err != nil {
		return nil, err
	}

	// The content is Base64 encoded, so it needs to be decoded
	decodedContent, err := base64.StdEncoding.DecodeString(githubFile.Content)
	if err != nil {
		return nil, err
	}

	return decodedContent, nil
}

//func (r *Repo) GetWorkflowRun(ctx context.Context, repository string, runId int64) (GithubWorkflowRun, error) {
//	// Get a workflow run for the given repository and runId
//	var workflowRun GithubWorkflowRun
//	err := r.do(ctx, nil, &workflowRun, requestOptions{
//		method:      http.MethodGet,
//		path:        githubAPIURL + "/repos/" + repository + "/actions/runs/" + strconv.FormatInt(runId, 10),
//		contentType: "application/json",
//	})
//	if err != nil {
//		return GithubWorkflowRun{}, err
//	}
//
//	return workflowRun, nil
//}

func (r *Repo) getWorkflowFile(ctx context.Context, repository string, path string) (string, error) {
	// Get the content of the workflow file
	var githubFile githubFile
	err := r.do(ctx, nil, &githubFile, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository + "/contents/" + path,
		contentType: "application/vnd.github.VERSION.raw",
	})
	if err != nil {
		return "", err
	}

	// The content is Base64 encoded, so it needs to be decoded
	decodedContent, err := base64.StdEncoding.DecodeString(githubFile.Content)
	if err != nil {
		return "", err
	}

	return string(decodedContent), nil
}

func (r *Repo) GetWorkflowRunLogs(ctx context.Context, repository string, runId int64) (GithubWorkflowRunLogs, error) {
	// Get the logs for a given workflow run
	var workflowRunLogs GithubWorkflowRunLogs
	err := r.do(ctx, nil, &workflowRunLogs, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository + "/actions/runs/" + strconv.FormatInt(runId, 10) + "/logs",
		contentType: "application/json",
	})
	if err != nil {
		return GithubWorkflowRunLogs{}, err
	}

	return workflowRunLogs, nil
}

func (r *Repo) ReRunFailedJobs(ctx context.Context, repository string, runId int64) error {
	// Re-run failed jobs for a given workflow run
	err := r.do(ctx, nil, nil, requestOptions{
		method:      http.MethodPost,
		path:        githubAPIURL + "/repos/" + repository + "/actions/runs/" + strconv.FormatInt(runId, 10) + "/rerun-failed-jobs",
		contentType: "application/json",
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) ReRunWorkflow(ctx context.Context, repository string, runId int64) error {
	// Re-run a given workflow run
	err := r.do(ctx, nil, nil, requestOptions{
		method:      http.MethodPost,
		path:        githubAPIURL + "/repos/" + repository + "/actions/runs/" + strconv.FormatInt(runId, 10) + "/rerun",
		contentType: "application/json",
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) CancelWorkflow(ctx context.Context, repository string, runId int64) error {
	// Cancel a given workflow run
	err := r.do(ctx, nil, nil, requestOptions{
		method:      http.MethodPost,
		path:        githubAPIURL + "/repos/" + repository + "/actions/runs/" + strconv.FormatInt(runId, 10) + "/cancel",
		contentType: "application/json",
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) do(ctx context.Context, requestBody any, responseBody any, requestOptions requestOptions) error {
	// Construct the request URL
	reqURL, err := url.Parse(requestOptions.path)
	if err != nil {
		return err
	}

	// Add query parameters
	query := reqURL.Query()
	for key, value := range requestOptions.queryParams {
		query.Add(key, value)
	}
	reqURL.RawQuery = query.Encode()

	reqBody, err := parseRequestBody(requestOptions, requestBody)
	if err != nil {
		return err
	}

	// Create the HTTP request
	req, err := http.NewRequest(requestOptions.method, reqURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	if requestOptions.contentType != "" {
		req.Header.Set("Content-Type", requestOptions.contentType)
	}
	if requestOptions.accept != "" {
		req.Header.Set("Accept", requestOptions.accept)
	}
	req.Header.Set("Authorization", "Bearer "+r.githubToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req = req.WithContext(ctx)

	// Perform the HTTP request using the injected client
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var errorResponse struct {
		Message string `json:"message"`
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Decode the error response body
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return err
		}

		return errors.New(errorResponse.Message)
	}

	// Decode the response body
	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseRequestBody(requestOptions requestOptions, requestBody any) ([]byte, error) {
	var reqBody []byte

	if requestBody == nil {
		return reqBody, nil
	}

	// Marshal the request body to JSON if accept/content type is JSON
	if requestOptions.accept == "application/json" || requestOptions.contentType == "application/json" {
		reqBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to parse request body: %w", err)
		}

		return reqBody, nil
	}

	reqStr, ok := requestBody.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert request body to string: %v", requestBody)
	}

	reqBody = []byte(reqStr)

	return reqBody, nil
}

type requestOptions struct {
	method      string
	path        string
	contentType string
	accept      string
	queryParams map[string]string
}

type githubWorkflow struct {
	TotalCount int64      `json:"total_count"`
	Workflows  []Workflow `json:"workflows"`
}

type workflowFile struct {
	On map[string]interface{} `yaml:"on"`
}

type githubFile struct {
	Content string `json:"content"`
}
