package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewRepository() Repository {
	return New("")
}

func TestRepo_LatestVersion(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()

	t.Run("Get latest version", func(t *testing.T) {
		res, err := repo.LatestVersion(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})
}
