package version

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (v *version) do(ctx context.Context, requestBody any, responseBody any, requestOptions requestOptions) error {
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

	// Marshal the request body to JSON if accept/content type is JSON
	reqBody, err := requestBodyToJSON(requestBody, requestOptions)
	if err != nil {
		return err
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
	resp, err := v.client.Do(req)
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
		if err = json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
			return err
		}
	}

	return nil
}

func requestBodyToJSON(requestBody any, requestOptions requestOptions) ([]byte, error) {
	if requestBody != nil {
		if requestOptions.accept == "application/json" || requestOptions.contentType == "application/json" {
			return json.Marshal(requestBody)
		}

		return []byte(requestBody.(string)), nil
	}

	return []byte{}, nil
}

type requestOptions struct {
	method      string
	path        string
	contentType string
	accept      string
	queryParams map[string]string
}
