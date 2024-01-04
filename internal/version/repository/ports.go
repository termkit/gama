package repository

type Repository interface {
	CurrentVersion() string
	LatestVersion() (string, error)
}
