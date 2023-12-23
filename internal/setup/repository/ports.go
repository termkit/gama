package repository

import (
	"context"
)

type Repository interface {
	GetConfig(ctx context.Context) (Config, error)
	UpdateConfig(ctx context.Context, config Config) error
}
