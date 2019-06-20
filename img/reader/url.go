package reader

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Http struct {
	// Headers will set headers on each request
	Headers http.Header
}

func (r *Http) Read(url string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	for k, v := range r.Headers {
		for _, headerVal := range v {
			req.Header.Add(k, headerVal)
		}
	}

	resp, err := http.DefaultClient.Do(req)
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
