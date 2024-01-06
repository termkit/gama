package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Repo struct {
	Client HttpClient

	currentVersion string
	latestVersion  string
}

const (
	owner = "termkit"
	repo  = "gama"
)

func New(currentVersion string) *Repo {
	return &Repo{
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
		currentVersion: currentVersion,
	}
}

func (r *Repo) CurrentVersion() string {
	return r.currentVersion
}

func (r *Repo) LatestVersion() (string, error) {
	var result struct {
		TagName string `json:"tag_name"`
	}

	err := r.do(context.Background(), nil, &result, requestOptions{
		method: "GET",
		path:   fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo),
		accept: "application/vnd.github+json",
	})
	// client time out error
	if errors.As(err, &context.DeadlineExceeded) {
		return "", errors.New("request timed out")
	} else if err != nil {
		return "", err
	}

	return result.TagName, nil
}

func (r *Repo) do(ctx context.Context, requestBody any, responseBody any, requestOptions requestOptions) error {
	// Construct the request URL
	reqURL, err := url.Parse(requestOptions.path)
	if err != nil {
		return err
	}

	// Add query parameters
	query := reqURL.Query()
	for key, value := range requestOptions.queryParams {
		query.Add(key, value)
	}
	reqURL.RawQuery = query.Encode()

	var reqBody []byte
	// Marshal the request body to JSON if accept/content type is JSON
	if requestOptions.accept == "application/json" || requestOptions.contentType == "application/json" {
		if requestBody != nil {
			reqBody, err = json.Marshal(requestBody)
			if err != nil {
				return err
			}
		}
	} else {
		if requestBody != nil {
			reqBody = []byte(requestBody.(string))
		}
	}

	// Create the HTTP request
	req, err := http.NewRequest(requestOptions.method, reqURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	if requestOptions.contentType == "" {
		req.Header.Set("Content-Type", requestOptions.contentType)
	}
	if requestOptions.accept == "" {
		req.Header.Set("Accept", requestOptions.accept)
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req = req.WithContext(ctx)

	// Perform the HTTP request using the injected client
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var errorResponse struct {
		Message string `json:"message"`
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Decode the error response body
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return err
		}

		return errors.New(errorResponse.Message)
	}

	// Decode the response body
	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
		if err != nil {
			return err
		}
	}

	return nil
}

type requestOptions struct {
	method      string
	path        string
	contentType string
	accept      string
	queryParams map[string]string
}
