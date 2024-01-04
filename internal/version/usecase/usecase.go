package usecase

import (
	"github.com/Masterminds/semver/v3"

	vr "github.com/termkit/gama/internal/version/repository"
)

type useCase struct {
	versionRepository vr.Repository
}

func New(repository vr.Repository) UseCase {
	return &useCase{
		versionRepository: repository,
	}
}

func (u *useCase) CurrentVersion() string {
	return u.versionRepository.CurrentVersion()
}

func (u *useCase) IsUpdateAvailable() (isAvailable bool, version string, err error) {
	currentVersion := u.versionRepository.CurrentVersion()
	if currentVersion == "under development" {
		return false, currentVersion, nil
	}

	latestVersion, err := u.versionRepository.LatestVersion()
	if err != nil {
		return false, currentVersion, err
	}

	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false, currentVersion, err
	}

	latest, err := semver.NewVersion(latestVersion)
	if err != nil {
		return false, currentVersion, err
	}

	return latest.GreaterThan(current), latestVersion, nil
}
