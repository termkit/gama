package repository

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// JoinPath joins URL host and paths with by removing all leading slashes from paths and adds a trailing slash to end of paths except last one.
func JoinPath(paths ...string) (*url.URL, error) {
	var uri = new(url.URL)
	for i, path := range paths {
		path = strings.TrimLeft(path, "/")
		if i+1 != len(paths) && !strings.HasSuffix(path, "/") {
			path = fmt.Sprintf("%s/", path)
		}

		u, err := url.Parse(path)
		if err != nil {
			return nil, err
		}

		uri = uri.ResolveReference(u)
	}

	return uri, nil
}
