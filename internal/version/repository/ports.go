package repository

import "context"

type Repository interface {
	CurrentVersion() string
	LatestVersion(ctx context.Context) (string, error)
}
