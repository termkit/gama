package usecase

import "context"

type UseCase interface {
	CurrentVersion() string
	IsUpdateAvailable(ctx context.Context) (isAvailable bool, version string, err error)
}
