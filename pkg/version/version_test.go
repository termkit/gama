package version

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	repositoryOwner    = "termkit"
	repositoryName     = "gama"
	testCurrentVersion = "1.0.0"
)

func NewRepository() Version {
	return New(repositoryOwner, repositoryName, testCurrentVersion)
}

func TestVersion_CurrentVersion(t *testing.T) {
	repo := NewRepository()
	t.Run("Get current version", func(t *testing.T) {
		assert.Equal(t, testCurrentVersion, repo.CurrentVersion())
	})
}

func TestVersion_IsUpdateAvailable(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()

	t.Run("Check if update is available", func(t *testing.T) {
		isAvailable, version, err := repo.IsUpdateAvailable(ctx)
		assert.NoError(t, err)
		assert.True(t, isAvailable)
		assert.NotEmpty(t, version)
	})
}

func TestRepo_LatestVersion(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()

	t.Run("Get latest version", func(t *testing.T) {
		res, err := repo.LatestVersion(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
		assert.NotEqual(t, testCurrentVersion, res)
	})
}
