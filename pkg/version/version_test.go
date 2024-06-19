package version

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	repositoryOwner = "termkit"
	repositoryName  = "gama"
)

func TestVersion_CurrentVersion(t *testing.T) {
	testCurrentVersion := "1.0.0"

	repo := New(repositoryOwner, repositoryName, testCurrentVersion)
	t.Run("Get current version", func(t *testing.T) {
		assert.Equal(t, testCurrentVersion, repo.CurrentVersion())
	})
}

func TestVersion_IsUpdateAvailable(t *testing.T) {
	ctx := context.Background()

	testCurrentVersion := "1.0.0"

	repo := New(repositoryOwner, repositoryName, testCurrentVersion)

	t.Run("Check if update is available", func(t *testing.T) {
		isAvailable, version, err := repo.IsUpdateAvailable(ctx)
		assert.NoError(t, err)
		assert.True(t, isAvailable)
		assert.NotEmpty(t, version)
	})
}

func TestVersion_Changelogs(t *testing.T) {
	ctx := context.Background()

	testCurrentVersion := "1.0.0"

	repo := New(repositoryOwner, repositoryName, testCurrentVersion)

	t.Run("Get changelogs", func(t *testing.T) {
		res, err := repo.Changelogs(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})
}

func TestVersion_ChangelogsSinceCurrentVersion(t *testing.T) {
	ctx := context.Background()

	testCurrentVersion := "1.1.0"

	repo := New(repositoryOwner, repositoryName, testCurrentVersion)

	t.Run("Get changelogs since current version", func(t *testing.T) {
		res, err := repo.ChangelogsSinceCurrentVersion(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})

}

func TestRepo_LatestVersion(t *testing.T) {
	ctx := context.Background()

	testCurrentVersion := "1.0.0"

	repo := New(repositoryOwner, repositoryName, testCurrentVersion)

	t.Run("Get latest version", func(t *testing.T) {
		res, err := repo.LatestVersion(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
		assert.NotEqual(t, testCurrentVersion, res)
	})
}
