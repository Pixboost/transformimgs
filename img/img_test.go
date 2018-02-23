package img_test

import (
	"errors"
	"github.com/dooman87/kolibri/test"
	"github.com/Pixboost/transformimgs/img"
	"net/http"
	"net/http/httptest"
	"testing"
	"net/url"
)

type resizerMock struct{}

func (r *resizerMock) Resize(data []byte, size string, imgId string) ([]byte, error) {
	if string(data) == "321" && size == "300x200" {
		return []byte("123"), nil
	}
	return nil, errors.New("resize_error")
}

func (r *resizerMock) FitToSize(data []byte, size string, imgId string) ([]byte, error) {
	if string(data) == "321" && size == "300x200" {
		return []byte("123"), nil
	}
	return nil, errors.New("resize_error")
}

func (r *resizerMock) Optimise(data []byte, imgId string) ([]byte, error) {
	if string(data) == "321" {
		return []byte("123"), nil
	}
	return nil, errors.New("resize_error")
}

type readerMock struct{}

func (r *readerMock) Read(url string) ([]byte, error) {
	if url == "http://site.com/img.png" {
		return []byte("321"), nil
	}
	return nil, errors.New("read_error")
}

func TestService_ResizeUrl(t *testing.T) {
	test.Service = createService(t).GetRouter().ServeHTTP
	test.T = t


	testCases := []test.TestCase{
		{
			Url:         "http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300x200",
			Description: "Success",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/resize?size=300x200", t),
				Header: map[string][]string {
					"X-Forwarded-Proto": {"http"},
				},
			},
			Description: "X-Forwarded-Proto",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/resize",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param size is required",
		},
		{
			Url:          "http://localhost/img/NO_SUCH_IMAGE/resize?size=300x200",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Read error",
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=BADSIZE",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Resize error",
		},
	}

	test.RunRequests(testCases)
}

func TestService_FitToSizeUrl(t *testing.T) {
	test.Service = createService(t).GetRouter().ServeHTTP
	test.T = t

	testCases := []test.TestCase{
		{
			Url:         "http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=300x200",
			Description: "Success",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/fit?size=300x200", t),
				Header: map[string][]string {
					"X-Forwarded-Proto": {"http"},
				},
			},
			Description: "X-Forwarded-Proto",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param size is required",
		},
		{
			Url:          "http://localhost/img/NO_SUCH_IMAGE/fit?size=300x200",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Read error",
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=BADSIZE",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Size param should be in format WxH",
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=300",
			ExpectedCode: http.StatusBadRequest,
			Description:  "2 - Size param should be in format WxH",
		},
	}

	test.RunRequests(testCases)
}

func TestService_OptimiseUrl(t *testing.T) {
	test.Service = createService(t).GetRouter().ServeHTTP
	test.T = t

	testCases := []test.TestCase{
		{
			Url:         "http://localhost/img/http%3A%2F%2Fsite.com/img.png/optimise",
			Description: "Success",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/optimise", t),
				Header: map[string][]string {
					"X-Forwarded-Proto": {"http"},
				},
			},
			Description: "X-Forwarded-Proto",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Url:          "http://localhost/img/NO_SUCH_IMAGE/optimise",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Read error",
		},
	}

	test.RunRequests(testCases)
}

func TestService_AsIs(t *testing.T) {
	test.Service = createService(t).GetRouter().ServeHTTP
	test.T = t

	testCases := []test.TestCase{
		{
			Url:         "http://localhost/img/http%3A%2F%2Fsite.com/img.png/asis",
			Description: "Success",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/asis", t),
				Header: map[string][]string {
					"X-Forwarded-Proto": {"http"},
				},
			},
			Description: "X-Forwarded-Proto",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Url:          "http://localhost/img/NO_SUCH_IMAGE/asis",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Read error",
		},
	}

	test.RunRequests(testCases)
}

func createService(t *testing.T) *img.Service {
	img.CacheTTL = 86400
	s, err := img.NewService(&readerMock{}, &resizerMock{}, 1)
	if err != nil {
		t.Fatalf("Error while creating service: %+v", err)
		return nil
	}
	return s
}

func parseUrl(strUrl string, t *testing.T) *url.URL {
	u, err := url.Parse(strUrl)
	if err != nil {
		t.Fatalf("Error while creating URL from [%s]: %v", strUrl, err)
		return nil
	}
	return u
}
