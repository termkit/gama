package repository

import (
	"context"
	"testing"
)

func NewRepository() Repository {
	return New("")
}

func TestRepo_LatestVersion(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()

	_, err := repo.LatestVersion(ctx)
	if err != nil {
		t.Error(err)
	}
}
