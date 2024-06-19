package version

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Masterminds/semver/v3"
)

type version struct {
	client HttpClient

	repositoryOwner string
	repositoryName  string

	currentVersion string
	latestVersion  string
}

// Changelog represents a changelog information
type Changelog struct {
	// PublishedAt is the date when the changelog was published
	PublishedAt time.Time `json:"published_at"`

	// TagName is the tag name of the changelog
	TagName string `json:"tag_name"`

	// Body message of the changelog
	Body string `json:"body"`
}

// New creates a new version instance
func New(repositoryOwner, repositoryName, currentVersion string) Version {
	return &version{
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
		repositoryOwner: repositoryOwner,
		repositoryName:  repositoryName,
		currentVersion:  currentVersion,
	}
}

// CurrentVersion returns the current version of the repository
func (v *version) CurrentVersion() string {
	return v.currentVersion
}

// LatestVersion returns the latest version of the repository
func (v *version) LatestVersion(ctx context.Context) (string, error) {
	var result struct {
		TagName string `json:"tag_name"`
	}

	err := v.do(ctx, nil, &result, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", v.repositoryOwner, v.repositoryName),
		accept: "application/vnd.github+json",
	})
	// client time out error
	var deadlineExceededError *url.Error
	if err != nil {
		if errors.As(err, &deadlineExceededError) && deadlineExceededError.Timeout() {
			return "", errors.New("request timed out")
		}
		return "", err
	}

	return result.TagName, nil
}

// Changelogs returns all changelogs of the repository
func (v *version) Changelogs(ctx context.Context) ([]Changelog, error) {
	var result []Changelog

	err := v.do(ctx, nil, &result, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", v.repositoryOwner, v.repositoryName),
		accept: "application/vnd.github+json",
	})
	// client time out error
	var deadlineExceededError *url.Error
	if err != nil {
		if errors.As(err, &deadlineExceededError) && deadlineExceededError.Timeout() {
			return nil, errors.New("request timed out")
		}
		return nil, err
	}

	return result, nil
}

// ChangelogsSinceCurrentVersion returns all changelogs since the current version
func (v *version) ChangelogsSinceCurrentVersion(ctx context.Context) ([]Changelog, error) {
	changelogs, err := v.Changelogs(ctx)
	if err != nil {
		return nil, err
	}

	currentVersion, err := semver.NewVersion(v.CurrentVersion())
	if err != nil {
		return nil, err
	}

	var result []Changelog
	for _, cl := range changelogs {
		tagVersion, err := semver.NewVersion(cl.TagName)
		if err != nil {
			continue
		}

		if tagVersion.GreaterThan(currentVersion) {
			result = append(result, cl)
		}
	}

	return result, nil
}

// IsUpdateAvailable checks if an update is available
func (v *version) IsUpdateAvailable(ctx context.Context) (isAvailable bool, version string, err error) {
	currentVersion := v.CurrentVersion()
	if currentVersion == "under development" {
		return false, currentVersion, nil
	}

	latestVersion, err := v.LatestVersion(ctx)
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
