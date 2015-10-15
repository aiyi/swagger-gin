package swag

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// LoadFromFileOrHTTP loads the bytes from a file or a remote http server based on the path passed in
func LoadFromFileOrHTTP(path string) ([]byte, error) {
	return LoadStrategy(path, ioutil.ReadFile, loadHTTPBytes)(path)
}

// LoadStrategy returns a loader function for a given path or uri
func LoadStrategy(path string, local, remote func(string) ([]byte, error)) func(string) ([]byte, error) {
	if strings.HasPrefix(path, "http") {
		return remote
	}
	return local
}

func loadHTTPBytes(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not access document at %q [%s] ", path, resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}
