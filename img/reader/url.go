package reader

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Http struct{}

func (r *Http) Read(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("Expected %d but got code %d.\n Error '%s'",
			http.StatusOK, resp.StatusCode, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return result, contentType, nil
}
