package loader

import (
	"context"
	"fmt"
	"github.com/Pixboost/transformimgs/v8/img"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Http struct {
	// Headers that will be sent with each request
	Headers http.Header
}

var dialer = &net.Dialer{
	Timeout:   5 * time.Second,
	KeepAlive: 30 * time.Second,
}

var client = &http.Client{
	Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func (r *Http) Load(url string, ctx context.Context) (*img.Image, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range r.Headers {
		for _, headerVal := range v {
			req.Header.Add(k, headerVal)
		}
	}

	if moreHeaders, ok := img.HeaderFromContext(ctx); ok {
		for k, v := range *moreHeaders {
			for _, headerVal := range v {
				req.Header.Add(k, headerVal)
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Expected %d but got code %d.\n Error '%s'",
			http.StatusOK, resp.StatusCode, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	contentEncoding := resp.Header.Get("Content-Encoding")

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &img.Image{
		Id:              url,
		Data:            result,
		MimeType:        contentType,
		ContentEncoding: contentEncoding,
	}, nil
}
