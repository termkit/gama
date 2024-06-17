package version

import "context"

type Version interface {
	CurrentVersion() string
	LatestVersion(ctx context.Context) (string, error)
	IsUpdateAvailable(ctx context.Context) (isAvailable bool, version string, err error)
}
