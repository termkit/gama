package repository

import (
	"context"
	"github.com/termkit/gama/internal/config"
	"reflect"
	"testing"

	"github.com/termkit/gama/internal/github/domain"
)

func newRepo(ctx context.Context) *Repo {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	repo := New(cfg)
	return repo
}

func TestRepo_ListRepositories(t *testing.T) {
	ctx := context.Background()

	repo := newRepo(ctx)

	limit := 30
	page := 1
	sort := domain.SortByUpdated

	repositories, err := repo.ListRepositories(ctx, limit, page, sort)
	if err != nil {
		t.Error(err)
	}

	if len(repositories) == 0 {
		t.Error("Expected repositories, got none")
	}
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

func TestRepo_GetAuthUser(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *GithubUser
		wantErr bool
	}{
		{
			name: "correct",

			args: args{
				ctx: context.Background(),
			},
			want:    &GithubUser{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r = newRepo(tt.args.ctx)
			got, err := r.GetAuthUser(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.GetAuthUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID == 0 {
				t.Errorf("Repo.GetAuthUser() nothing come: %v, wantErr %v", got, tt.wantErr)
				return
			}
		})
	}
}

func TestRepo_ListBranches(t *testing.T) {
	type args struct {
		ctx        context.Context
		repository string
	}
	tests := []struct {
		name    string
		args    args
		want    []GithubBranch
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				ctx:        context.Background(),
				repository: "fleimkeipa/dvpwa", // public repo
			},
			want: []GithubBranch{
				{
					Name: "develop",
				},
				{
					Name: "imported",
				},
				{
					Name: "main",
				},
				{
					Name: "master",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r = newRepo(tt.args.ctx)
			got, err := r.ListBranches(tt.args.ctx, tt.args.repository)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.ListBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repo.ListBranches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_GetRepository(t *testing.T) {
	type args struct {
		ctx        context.Context
		repository string
	}
	tests := []struct {
		name    string
		args    args
		want    *GithubRepository
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				ctx:        context.Background(),
				repository: "fleimkeipa/dvpwa", // public repo
			},
			want: &GithubRepository{
				Name: "dvpwa",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r = newRepo(tt.args.ctx)
			got, err := r.GetRepository(tt.args.ctx, tt.args.repository)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.GetRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("Repo.GetRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_GetWorkflows(t *testing.T) {
	type args struct {
		ctx        context.Context
		repository string
	}
	tests := []struct {
		name    string
		args    args
		want    []Workflow
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				ctx:        context.Background(),
				repository: "fleimkeipa/dvpwa", // public repo
			},
			want:    []Workflow{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r = newRepo(tt.args.ctx)
			got, err := r.GetWorkflows(tt.args.ctx, tt.args.repository)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.GetWorkflows() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("Repo.GetWorkflows() = %v, want %v", len(got), 2)
			}
		})
	}
}

func TestRepo_InspectWorkflowContent(t *testing.T) {
	type args struct {
		ctx          context.Context
		repository   string
		branch       string
		workflowFile string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				ctx:          context.Background(),
				repository:   "fleimkeipa/dvpwa",
				branch:       "master",
				workflowFile: ".github/workflows/build.yml",
			},
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r = newRepo(tt.args.ctx)
			got, err := r.InspectWorkflowContent(tt.args.ctx, tt.args.repository, tt.args.branch, tt.args.workflowFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.InspectWorkflowContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("Repo.GetWorkflows() = %v, want %v", len(got), 2)
			}
		})
	}
}
