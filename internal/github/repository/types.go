package repository

import (
	"time"
)

type GithubConfig struct {
	Token string
}

type InitializeOptions struct {
	HTTPTimeout time.Duration
}

type GithubRepository struct {
	Id              int         `json:"id"`
	NodeId          string      `json:"node_id"`
	Name            string      `json:"name"`
	FullName        string      `json:"full_name"`
	Private         bool        `json:"private"`
	Description     string      `json:"description"`
	Language        interface{} `json:"language"`
	ForksCount      int         `json:"forks_count"`
	StargazersCount int         `json:"stargazers_count"`
	WatchersCount   int         `json:"watchers_count"`
	Size            int         `json:"size"`
	DefaultBranch   string      `json:"default_branch"`
	OpenIssuesCount int         `json:"open_issues_count"`
	IsTemplate      bool        `json:"is_template"`
	Topics          []string    `json:"topics"`
	HasIssues       bool        `json:"has_issues"`
	HasProjects     bool        `json:"has_projects"`
	HasWiki         bool        `json:"has_wiki"`
	HasPages        bool        `json:"has_pages"`
	HasDownloads    bool        `json:"has_downloads"`
	Archived        bool        `json:"archived"`
	Disabled        bool        `json:"disabled"`
	Visibility      string      `json:"visibility"`
	PushedAt        time.Time   `json:"pushed_at"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Permissions     struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	} `json:"permissions"`
	AllowRebaseMerge    bool        `json:"allow_rebase_merge"`
	TemplateRepository  interface{} `json:"template_repository"`
	TempCloneToken      string      `json:"temp_clone_token"`
	AllowSquashMerge    bool        `json:"allow_squash_merge"`
	AllowAutoMerge      bool        `json:"allow_auto_merge"`
	DeleteBranchOnMerge bool        `json:"delete_branch_on_merge"`
	AllowMergeCommit    bool        `json:"allow_merge_commit"`
	SubscribersCount    int         `json:"subscribers_count"`
	NetworkCount        int         `json:"network_count"`
	License             struct {
		Key     string `json:"key"`
		Name    string `json:"name"`
		Url     string `json:"url"`
		SpdxId  string `json:"spdx_id"`
		NodeId  string `json:"node_id"`
		HtmlUrl string `json:"html_url"`
	} `json:"license"`
	Forks      int `json:"forks"`
	OpenIssues int `json:"open_issues"`
	Watchers   int `json:"watchers"`
}

type GithubBranch struct {
	Name string `json:"name"`
}

type Workflow struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
	Url       string    `json:"url"`
	HtmlUrl   string    `json:"html_url"`
}

type GithubWorkflowRun struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

type WorkflowRun struct {
	Id          int64  `json:"id"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	HeadBranch  string `json:"head_branch"`

	Actor Actor `json:"actor"`

	URL       string `json:"url"`
	JobsURL   string `json:"jobs_url"`
	LogsURL   string `json:"logs_url"`
	CancelURL string `json:"cancel_url"`
	RerunURL  string `json:"rerun_url"`

	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	RunAttempt int       `json:"run_attempt"`
}

type Actor struct {
	Id        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
}

type GithubWorkflowRunLogs struct {
	TotalSize int    `json:"total_size"`
	Url       string `json:"url"`
	Download  string `json:"download_url"`
}
