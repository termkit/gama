package repository

import (
	"testing"
)

func NewRepository() Repository {
	return New("")
}

func TestRepo_LatestVersion(t *testing.T) {
	repo := NewRepository()

	_, err := repo.LatestVersion()
	if err != nil {
		t.Error(err)
	}
}
