package img_test

import (
	"errors"
	"github.com/dooman87/kolibri/test"
	"github.com/dooman87/transformimgs/img"
	"net/http"
	"net/http/httptest"
	"testing"
)

type resizerMock struct{}

func (r *resizerMock) Resize(data []byte, size string) ([]byte, error) {
	if string(data) == "321" && size == "300x200" {
		return []byte("123"), nil
	}
	return nil, errors.New("resize_error")
}

func (r *resizerMock) FitToSize(data []byte, size string) ([]byte, error) {
	if string(data) == "321" && size == "300x200" {
		return []byte("123"), nil
	}
	return nil, errors.New("resize_error")
}

func (r *resizerMock) Optimise(data []byte) ([]byte, error) {
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
	test.Service = createService(t).ResizeUrl
	test.T = t

	testCases := []test.TestCase{
		{
			Url:         "http://localhost/img?url=http://site.com/img.png&size=300x200",
			Description: "Success",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Url:          "http://localhost/img?size=300x200",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param url is required",
		},
		{
			Url:          "http://localhost/img?url=http://site.com/img.png",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param size is required",
		},
		{
			Url:          "http://localhost/img?url=NO_SUCH_IMAGE&size=300x200",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Read error",
		},
		{
			Url:          "http://localhost/img?url=http://site.com/img.png&size=BADSIZE",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Resize error",
		},
	}

	test.RunRequests(testCases)
}

func TestService_FitToSizeUrl(t *testing.T) {
	test.Service = createService(t).FitToSizeUrl
	test.T = t

	testCases := []test.TestCase{
		{
			Url:         "http://localhost/fit?url=http://site.com/img.png&size=300x200",
			Description: "Success",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Url:          "http://localhost/fit?size=300x200",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param url is required",
		},
		{
			Url:          "http://localhost/fit?url=http://site.com/img.png",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param size is required",
		},
		{
			Url:          "http://localhost/fit?url=NO_SUCH_IMAGE&size=300x200",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Read error",
		},
		{
			Url:          "http://localhost/fit?url=http://site.com/img.png&size=BADSIZE",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Size param should be in format WxH",
		},
		{
			Url:          "http://localhost/fit?url=http://site.com/img.png&size=300",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Size param should be in format WxH",
		},
	}

	test.RunRequests(testCases)
}

func TestService_OptimiseUrl(t *testing.T) {
	test.Service = createService(t).OptimiseUrl
	test.T = t

	testCases := []test.TestCase{
		{
			Url:         "http://localhost/img?url=http://site.com/img.png",
			Description: "Success",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Url:          "http://localhost/img",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param url is required",
		},
		{
			Url:          "http://localhost/fit?url=NO_SUCH_IMAGE",
			ExpectedCode: http.StatusInternalServerError,
			Description:  "Read error",
		},
	}

	test.RunRequests(testCases)
}

func TestService_AsIs(t *testing.T) {
	test.Service = createService(t).AsIs
	test.T = t

	testCases := []test.TestCase{
		{
			Url:         "http://localhost/asis?url=http://site.com/img.png",
			Description: "Success",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
				)
			},
		},
		{
			Url:          "http://localhost/img",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param url is required",
		},
		{
			Url:          "http://localhost/fit?url=NO_SUCH_IMAGE",
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
