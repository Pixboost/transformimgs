package img_test

import (
	"context"
	"errors"
	"github.com/Pixboost/transformimgs/v6/img"
	"github.com/dooman87/kolibri/test"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	NoContentTypeImgSrc = "111"
	NoContentTypeImgOut = "222"

	ImgSrc           = "321"
	ImgAvifOut       = "12345"
	ImgWebpOut       = "1234"
	ImgPngOut        = "123"
	ImgLowQualityOut = "12"
)

type resizerMock struct{}

func (r *resizerMock) Resize(config *img.TransformationConfig) (*img.Image, error) {
	data := config.Src.Data
	size := config.Config.(*img.ResizeConfig).Size
	if (string(data) != ImgSrc && string(data) != NoContentTypeImgSrc) || size != "300x200" {
		return nil, errors.New("resize_error")
	}

	return r.resultImage(config), nil
}

func (r *resizerMock) FitToSize(config *img.TransformationConfig) (*img.Image, error) {
	data := config.Src.Data
	size := config.Config.(*img.ResizeConfig).Size
	if (string(data) != ImgSrc && string(data) != NoContentTypeImgSrc) || size != "300x200" {
		return nil, errors.New("fit_error")
	}

	return r.resultImage(config), nil
}

func (r *resizerMock) Optimise(config *img.TransformationConfig) (*img.Image, error) {
	data := config.Src.Data

	if string(data) != ImgSrc && string(data) != NoContentTypeImgSrc {
		return nil, errors.New("optimise_error")
	}

	return r.resultImage(config), nil
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

func (r *resizerMock) resultImage(config *img.TransformationConfig) *img.Image {
	if string(config.Src.Data) == NoContentTypeImgSrc {
		return &img.Image{
			Data: []byte(NoContentTypeImgOut),
		}
	}

	if config.Quality == img.LOW {
		return &img.Image{
			Data: []byte(ImgLowQualityOut),
		}
	}

	if r.supports(config.SupportedFormats, "image/avif") {
		return &img.Image{
			Data:     []byte(ImgAvifOut),
			MimeType: "image/avif",
		}
	}

	if r.supports(config.SupportedFormats, "image/webp") {
		return &img.Image{
			Data:     []byte(ImgWebpOut),
			MimeType: "image/webp",
		}
	}

	return &img.Image{
		Data:     []byte(ImgPngOut),
		MimeType: "image/png",
	}
}

type loaderMock struct{}

func (l *loaderMock) Load(url string, ctx context.Context) (*img.Image, error) {
	if url == "http://site.com/img.png" {
		return &img.Image{
			Data:     []byte(ImgSrc),
			MimeType: "image/png",
		}, nil
	}
	if url == "http://site.com/img2.png" {
		return &img.Image{
			Data:     []byte(NoContentTypeImgSrc),
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
			Description: "Success",
			Url:         "http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300x200",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
					test.Equal("Accept, Save-Data", w.Header().Get("Vary"), "Vary header"),
					test.Equal("image/png", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "WEBP Support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300x200", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("4", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgWebpOut, w.Body.String(), "Resulted image"),
					test.Equal("image/webp", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "AVIF Support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300x200", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp, image/avif"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("5", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgAvifOut, w.Body.String(), "Resulted image"),
					test.Equal("image/avif", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "Save-Data support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300x200", t),
				Header: map[string][]string{
					"Save-Data": {"on"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("2", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgLowQualityOut, w.Body.String(), "Resulted image"),
				)
			},
		},
		{
			Description: "MIME Sniffing",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img2.png/resize?size=300x200", t),
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("text/plain; charset=utf-8", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "Using protocol from X-Forwarded-Proto header to load source image",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/resize?size=300x200", t),
				Header: map[string][]string{
					"X-Forwarded-Proto": {"http"},
				},
			},
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
			ExpectedCode: http.StatusBadRequest,
			Description:  "Resize error",
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300xx",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Resize error",
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=abcx200",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Resize error",
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/resize?size=300xabc",
			ExpectedCode: http.StatusBadRequest,
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
			Description: "Success",
			Url:         "http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=300x200",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal("Accept, Save-Data", w.Header().Get("Vary"), "Vary header"),
					test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
					test.Equal("image/png", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "WebP Support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=300x200", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("4", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgWebpOut, w.Body.String(), "Resulted image"),
					test.Equal("image/webp", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "AVIF Support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=300x200", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp, image/avif"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("5", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgAvifOut, w.Body.String(), "Resulted image"),
					test.Equal("image/avif", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "Save-Data support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?size=300x200", t),
				Header: map[string][]string{
					"Save-Data": {"on"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("2", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgLowQualityOut, w.Body.String(), "Resulted image"),
				)
			},
		},
		{
			Description: "MIME Sniffing",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img2.png/fit?size=300x200", t),
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("text/plain; charset=utf-8", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "Using protocol from X-Forwarded-Proto header to load source image",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/fit?size=300x200", t),
				Header: map[string][]string{
					"X-Forwarded-Proto": {"http"},
				},
			},
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
			Description: "Success",
			Url:         "http://localhost/img/http%3A%2F%2Fsite.com/img.png/optimise",
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("public, max-age=86400", w.Header().Get("Cache-Control"), "Cache-Control header"),
					test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal("Accept, Save-Data", w.Header().Get("Vary"), "Vary header"),
					test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
					test.Equal("image/png", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "Webp Support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/optimise", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("4", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgWebpOut, w.Body.String(), "Resulted image"),
					test.Equal("image/webp", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "AVIF Support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/optimise", t),
				Header: map[string][]string{
					"Accept": {"image/png, image/webp, image/avif"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("5", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgAvifOut, w.Body.String(), "Resulted image"),
					test.Equal("image/avif", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "Save-Data support",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img.png/optimise", t),
				Header: map[string][]string{
					"Save-Data": {"on"},
				},
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("2", w.Header().Get("Content-Length"), "Content-Length header"),
					test.Equal(ImgLowQualityOut, w.Body.String(), "Resulted image"),
				)
			},
		},
		{
			Description: "MIME Sniffing",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/http%3A%2F%2Fsite.com/img2.png/optimise", t),
			},
			Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
				test.Error(t,
					test.Equal("text/plain; charset=utf-8", w.Header().Get("Content-Type"), "Content-Type header"),
				)
			},
		},
		{
			Description: "Using protocol from X-Forwarded-Proto header to load source image",
			Request: &http.Request{
				Method: "GET",
				URL:    parseUrl("http://localhost/img/%2F%2Fsite.com/img.png/optimise", t),
				Header: map[string][]string{
					"X-Forwarded-Proto": {"http"},
				},
			},
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
			Description: "Success",
			Url:         "http://localhost/img/http%3A%2F%2Fsite.com/img.png/asis",
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
