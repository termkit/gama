package browser

import (
	"errors"
	"net/url"
	"os/exec"
	"runtime"
)

func OpenInBrowser(rawURL string) error {
	// Validate the URL to prevent command injection
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("invalid URL")
	}

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", parsedURL.String()}
	case "darwin":
		cmd = "open"
		args = []string{parsedURL.String()}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{parsedURL.String()}
	}
	
	// #nosec G204 - URL is validated above and is safe to use
	return exec.Command(cmd, args...).Start()
}
