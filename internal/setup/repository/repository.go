package repository

import (
	"context"
)

type Repo struct {
}

func New() *Repo {
	return &Repo{}
}

func (r *Repo) GetConfig(ctx context.Context) (Config, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Repo) UpdateConfig(ctx context.Context, config Config) error {
	//TODO implement me
	panic("implement me")
}
