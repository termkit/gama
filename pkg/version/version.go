package version

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"net/http"
	"net/url"
	"time"
)

type version struct {
	client HttpClient

	repositoryOwner string
	repositoryName  string

	currentVersion string
	latestVersion  string
}

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

func (v *version) CurrentVersion() string {
	return v.currentVersion
}

func (v *version) LatestVersion(ctx context.Context) (string, error) {
	var result struct {
		TagName string `json:"tag_name"`
	}

	err := v.do(ctx, nil, &result, requestOptions{
		method: "GET",
		path:   fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", v.repositoryOwner, v.repositoryName),
		accept: "application/vnd.github+json",
	})
	// client time out error
	var deadlineExceededError *url.Error
	if errors.As(err, &deadlineExceededError) && deadlineExceededError.Timeout() {
		return "", errors.New("request timed out")
	} else if err != nil {
		return "", err
	}

	return result.TagName, nil
}

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
