package img

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type ImgUrlReader struct{}

func (r *ImgUrlReader) Read(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Expected %d but got code %d.\n Error '%s'",
			http.StatusOK, resp.StatusCode, resp.Status)
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return result, nil
}
