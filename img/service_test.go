package img_test

import (
	"context"
	"errors"
	"github.com/Pixboost/transformimgs/img"
	"github.com/dooman87/kolibri/test"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type resizerMock struct{}

func (r *resizerMock) Resize(data []byte, size string, imgId string, supportedFormats []string) (*img.Image, error) {
	if (string(data) != "321" && string(data) != "111") || size != "300x200" {
		return nil, errors.New("resize_error")
	}

	return r.resultImage(data, supportedFormats), nil
}

func (r *resizerMock) FitToSize(data []byte, size string, imgId string, supportedFormats []string) (*img.Image, error) {
	if (string(data) != "321" && string(data) != "111") || size != "300x200" {
		return nil, errors.New("fit_error")
	}

	return r.resultImage(data, supportedFormats), nil
}

func (r *resizerMock) Optimise(data []byte, imgId string, supportedFormats []string) (*img.Image, error) {
	if string(data) != "321" && string(data) != "111" {
		return nil, errors.New("optimise_error")
	}

	return r.resultImage(data, supportedFormats), nil
}

func (r *resizerMock) supports(supportedFormats []string, format string) bool {
	supports := false
	for _, f := range supportedFormats {
		if f == format {
			supports = true
		}
	}

	return supports
}

func (r *resizerMock) resultImage(data []byte, supportedFormats []string) *img.Image {
	if string(data) == "111" {
		return &img.Image{
			Data: []byte("222"),
		}
	}

	if r.supports(supportedFormats, "image/avif") {
		return &img.Image{
			Data:     []byte("12345"),
			MimeType: "image/avif",
		}
	}

	if r.supports(supportedFormats, "image/webp") {
		return &img.Image{
			Data:     []byte("1234"),
			MimeType: "image/webp",
		}
	}

	return &img.Image{
		Data:     []byte("123"),
		MimeType: "image/png",
	}
}

type loaderMock struct{}

func (l *loaderMock) Load(url string, ctx context.Context) (*img.Image, error) {
	if url == "http://site.com/img.png" {
		return &img.Image{
			Data:     []byte("321"),
			MimeType: "image/png",
		}, nil
	}
	if url == "http://site.com/img2.png" {
		return &img.Image{
			Data:     []byte("111"),
			MimeType: "image/png",
		}, nil
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
					test.Equal("123", w.Body.String(), "Resulted image"),
					test.Equal("Accept", w.Header().Get("Vary"), "Vary header"),
					test.Equal("image/png", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300x200", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp"},
				},
			},
			Description: "WEBP Support",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("4", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal("1234", w.Body.String(), "Resulted image"),
					test.Equal("image/webp", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300x200", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp, image/avif"},
				},
			},
			Description: "AVIF Support",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("5", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal("12345", w.Body.String(), "Resulted image"),
					test.Equal("image/avif", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img2.png/resize?size=300x200", t),
			},
			Description: "MIME Sniffing",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("text/plain; charset=utf-8", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/resize?size=300x200", t),
				Header: map[string][]string{
					"X-Forwarded-Proto": {"http"},
				},
			},
			Description: "Using protocol from X-Forwarded-Proto header to load source image",
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
					test.Equal("Accept", w.Header().Get("Vary"), "Vary header"),
					test.Equal("123", w.Body.String(), "Resulted image"),
					test.Equal("image/png", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=300x200", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp"},
				},
			},
			Description: "WebP Support",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("4", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal("1234", w.Body.String(), "Resulted image"),
					test.Equal("image/webp", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=300x200", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp, image/avif"},
				},
			},
			Description: "AVIF Support",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("5", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal("12345", w.Body.String(), "Resulted image"),
					test.Equal("image/avif", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img2.png/fit?size=300x200", t),
			},
			Description: "MIME Sniffing",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("text/plain; charset=utf-8", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/fit?size=300x200", t),
				Header: map[string][]string{
					"X-Forwarded-Proto": {"http"},
				},
			},
			Description: "Using protocol from X-Forwarded-Proto header to load source image",
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
					test.Equal("Accept", w.Header().Get("Vary"), "Vary header"),
					test.Equal("123", w.Body.String(), "Resulted image"),
					test.Equal("image/png", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/optimise", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp"},
				},
			},
			Description: "Webp Support",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("4", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal("1234", w.Body.String(), "Resulted image"),
					test.Equal("image/webp", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/optimise", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp, image/avif"},
				},
			},
			Description: "AVIF Support",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("5", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal("12345", w.Body.String(), "Resulted image"),
					test.Equal("image/avif", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img2.png/optimise", t),
			},
			Description: "MIME Sniffing",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("text/plain; charset=utf-8", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/optimise", t),
				Header: map[string][]string{
					"X-Forwarded-Proto": {"http"},
				},
			},
			Description: "Using protocol from X-Forwarded-Proto header to load source image",
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
					test.Equal("image/png", w.Header().Get("Content-Type"), "Content-Type header"),
					test.Equal("", w.Header().Get("Vary"), "No Vary header"),
				)
			},
		},
		{
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/asis", t),
				Header: map[string][]string{
					"X-Forwarded-Proto": {"http"},
				},
			},
			Description: "Using protocol from X-Forwarded-Proto header to load source image",
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
	s, err := img.NewService(&loaderMock{}, &resizerMock{}, 1)
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
