package usecase

import (
	"context"
)

type UseCase interface {
	Setup(ctx context.Context) error
}
