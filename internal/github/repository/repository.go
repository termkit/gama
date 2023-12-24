package repository

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Repo struct {
	Client HttpClient

	githubToken string
}

var githubAPIURL = "https://api.github.com"

func New() *Repo {
	return &Repo{
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (r *Repo) Initialize(ctx context.Context, config GithubConfig, opts ...InitializeOptions) {
	r.githubToken = config.Token
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

func (r *Repo) ListRepositories(ctx context.Context) ([]GithubRepository, error) {
	// List repositories for the authenticated user
	var repositories []GithubRepository
	err := r.do(ctx, nil, &repositories, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/user/repos",
		contentType: "application/json",
		queryParams: map[string]string{
			"visibility": "private",
		},
	})
	if err != nil {
		return nil, err
	}

	return repositories, nil
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
	// Trigger a workflow for the given repository and branch
	err := r.do(ctx, workflow, nil, requestOptions{
		method:      http.MethodPost,
		path:        githubAPIURL + "/repos/" + repository + "/actions/workflows/" + workflowName + "/dispatches",
		contentType: "application/json",
		queryParams: map[string]string{
			"ref": branch,
		},
	})
	if err != nil {
		return err
	}

	return nil
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

	// Filter workflows to only include those that are dispatchable and manually triggerable
	var triggerableWorkflows []Workflow
	for _, workflow := range workflows.Workflows {
		// Get the workflow file content
		fileContent, err := r.getWorkflowFile(ctx, repository, workflow.Path)
		if err != nil {
			return nil, err
		}

		// Parse the workflow file content as YAML
		var wfFile workflowFile
		err = yaml.Unmarshal([]byte(fileContent), &wfFile)
		if err != nil {
			return nil, err
		}

		// Check if the workflow file content has a "workflow_dispatch" key
		if _, ok := wfFile.On["workflow_dispatch"]; ok {
			triggerableWorkflows = append(triggerableWorkflows, workflow)
		}
	}

	return triggerableWorkflows, nil
}

func (r *Repo) InspectWorkflowContent(ctx context.Context, repository string, workflowFile string) ([]byte, error) {
	// Get the content of the workflow file
	var githubFile githubFile
	err := r.do(ctx, nil, &githubFile, requestOptions{
		method:      http.MethodGet,
		path:        githubAPIURL + "/repos/" + repository + "/contents/" + workflowFile,
		contentType: "application/vnd.github.VERSION.raw",
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

	// Marshal the request body to JSON
	var reqBody []byte
	if requestBody != nil {
		reqBody, err = json.Marshal(requestBody)
		if err != nil {
			return err
		}
	}

	// Create the HTTP request
	req, err := http.NewRequest(requestOptions.method, reqURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", requestOptions.contentType)
	req.Header.Set("Authorization", "Bearer "+r.githubToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req = req.WithContext(ctx)

	// Perform the HTTP request using the injected client
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode the response body
	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
		if err != nil {
			return err
		}
	}

	return nil
}

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

type requestOptions struct {
	method      string
	path        string
	contentType string
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
