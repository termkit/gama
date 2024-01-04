package usecase

type UseCase interface {
	CurrentVersion() string
	IsUpdateAvailable() (isAvailable bool, version string, err error)
}
