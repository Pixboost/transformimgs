package img_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Pixboost/transformimgs/v8/img"
	"github.com/dooman87/kolibri/test"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	NoContentTypeImgSrc = "111"
	NoContentTypeImgOut = "222"

	ImgSrc             = "321"
	ImgAvifOut         = "12345"
	ImgWebpOut         = "1234"
	ImgPngOut          = "123"
	ImgLowQualityOut   = "12"
	ImgLowerQualityOut = "1"
	ImgBorderTrimmed   = "777"

	EmptyGifBase64Out = "R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
)

type resizerMock struct {
	fuzzTests bool
}

func (r *resizerMock) Resize(config *img.TransformationConfig) (*img.Image, error) {
	if !r.fuzzTests {
		data := config.Src.Data
		size := config.Config.(*img.ResizeConfig).Size
		if (string(data) != ImgSrc && string(data) != NoContentTypeImgSrc) || size != "300x200" {
			return nil, errors.New("resize_error")
		}
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
	if config.TrimBorder {
		return &img.Image{
			Data: []byte(ImgBorderTrimmed),
		}
	}

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

	if config.Quality == img.LOWER {
		return &img.Image{
			Data: []byte(ImgLowerQualityOut),
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
			Id:       url,
		}, nil
	}
	if url == "http://site.com/img2.png" {
		return &img.Image{
			Data:     []byte(NoContentTypeImgSrc),
			MimeType: "image/png",
			Id:       url,
		}, nil
	}
	return nil, errors.New("read_error")
}

type transformTest struct {
	name      string
	urlSuffix string
}

func TestNewService(t *testing.T) {
	_, err := img.NewService(nil, nil, 0)

	if err == nil || err.Error() != "procNum must be positive, but got [0]" {
		t.Errorf("expected error but got %s", err)
	}
}

func TestService_Transforms(t *testing.T) {
	tests := []*transformTest{
		{
			name:      "Resize",
			urlSuffix: "/resize?size=300x200&",
		},
		{
			name:      "Fit",
			urlSuffix: "/fit?size=300x200&",
		},
		{
			name:      "Optimise",
			urlSuffix: "/optimise?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Service = createService(t).GetRouter().ServeHTTP
			test.T = t

			testCases := []test.TestCase{
				{
					Description: "Success",
					Url:         fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s", tt.urlSuffix),
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
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s", tt.urlSuffix), t),
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
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s", tt.urlSuffix), t),
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
					Description: "Save-Data: on",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s", tt.urlSuffix), t),
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
					Description: "?save-data=off",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&save-data=off", tt.urlSuffix), t),
						Header: map[string][]string{
							"Save-Data": {"on"},
						},
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
							test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
						)
					},
				},
				{
					Description: "?save-data=hide",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&save-data=hide", tt.urlSuffix), t),
						Header: map[string][]string{
							"Save-Data": {"on"},
						},
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("image/gif", w.Header().Get("Content-Type"), "Content-Type header"),
							test.Equal(EmptyGifBase64Out, base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(w.Body.Bytes()), "Resulted image"),
						)
					},
				},
				{
					Description: "Invalid save-data param",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&save-data=hello", tt.urlSuffix), t),
						Header: map[string][]string{
							"Save-Data": {"on"},
						},
					},
					ExpectedCode: http.StatusBadRequest,
				},
				{
					Description: "DPPX > 2",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&dppx=2.625", tt.urlSuffix), t),
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("1", w.Header().Get("Content-Length"), "Content-Length header"),
							test.Equal(ImgLowerQualityOut, w.Body.String(), "Resulted image"),
						)
					},
				},
				{
					Description: "DPPX < 2",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&dppx=1", tt.urlSuffix), t),
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
							test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
						)
					},
				},
				{
					Description: "Invalid dppx",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&dppx=abc", tt.urlSuffix), t),
					},
					ExpectedCode: http.StatusBadRequest,
				},
				{
					Description: "MIME Sniffing",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img2.png%s", tt.urlSuffix), t),
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
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/%%2F%%2Fsite.com/img.png%s", tt.urlSuffix), t),
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
					Description: "Trim Border",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&trim-border", tt.urlSuffix), t),
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
							test.Equal(ImgBorderTrimmed, w.Body.String(), "Resulted image"),
						)
					},
				},
				{
					Description: "Trim Border False",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&trim-border=0", tt.urlSuffix), t),
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
							test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
						)
					},
				},
				{
					Url:          fmt.Sprintf("http://localhost/img/NO_SUCH_IMAGE%s", tt.urlSuffix),
					ExpectedCode: http.StatusInternalServerError,
					Description:  "Read error",
				},
			}

			test.RunRequests(testCases)
		})
	}
}

func TestService_Transforms_SaveDataDisabled(t *testing.T) {
	img.SaveDataEnabled = false

	tests := []*transformTest{
		{
			name:      "Resize",
			urlSuffix: "/resize?size=300x200&",
		},
		{
			name:      "Fit",
			urlSuffix: "/fit?size=300x200&",
		},
		{
			name:      "Optimise",
			urlSuffix: "/optimise?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Service = createService(t).GetRouter().ServeHTTP
			test.T = t

			testCases := []test.TestCase{
				{
					Description: "Save-Data: on",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s", tt.urlSuffix), t),
						Header: map[string][]string{
							"Save-Data": {"on"},
						},
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
							test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
							test.Equal("Accept", w.Header().Get("Vary"), "Vary response header"),
						)
					},
				},
				{
					Description: "Save-Data: off",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&save-data=off", tt.urlSuffix), t),
						Header: map[string][]string{
							"Save-Data": {"on"},
						},
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
							test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
							test.Equal("Accept", w.Header().Get("Vary"), "Vary response header"),
						)
					},
				},
				{
					Description: "Save-Data: hide",
					Request: &http.Request{
						Method: "GET",
						URL:    parseUrl(fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png%s&save-data=hide", tt.urlSuffix), t),
						Header: map[string][]string{
							"Save-Data": {"on"},
						},
					},
					Handler: func(w *httptest.ResponseRecorder, t *testing.T) {
						test.Error(t,
							test.Equal("3", w.Header().Get("Content-Length"), "Content-Length header"),
							test.Equal(ImgPngOut, w.Body.String(), "Resulted image"),
							test.Equal("Accept", w.Header().Get("Vary"), "Vary response header"),
						)
					},
				},
			}

			test.RunRequests(testCases)
		})
	}

	img.SaveDataEnabled = true
}

func TestService_ResizeUrl(t *testing.T) {
	test.Service = createService(t).GetRouter().ServeHTTP
	test.T = t

	testCases := []test.TestCase{
		{
			Url:          "http://localhost/img//resize?size=30x30",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Source image URL is required",
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
			Url:          "http://localhost/img//fit?size=30x30",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Source image URL is required",
		},
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Param size is required",
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

func TestService_TrimBorder(t *testing.T) {
	test.Service = createService(t).GetRouter().ServeHTTP
	test.T = t

	testCases := []test.TestCase{
		{
			Url:          "http://localhost/img/http%3A%2F%2Fsite.com/img.png/fit?trim-border=abc",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Source image URL is required",
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
		{
			Url:          "http://localhost/img//asis",
			ExpectedCode: http.StatusBadRequest,
			Description:  "Source image URL is required",
		},
	}

	test.RunRequests(testCases)
}

func FuzzService_ResizeUrl(f *testing.F) {
	f.Add("300x200", "image/png, image/webp, image/avif", 3.0, true, "")
	f.Add("300", "image/png, image/webp", 4.2, false, "off")
	f.Add("x200", "image/png", 4.2, false, "hide")

	img.CacheTTL = 86400
	srv, _ := img.NewService(&loaderMock{}, &resizerMock{fuzzTests: true}, 1)
	s := srv.GetRouter().ServeHTTP

	f.Fuzz(func(t *testing.T, size string, acceptFormats string, dppx float64, saveDataHeaderEnabled bool, saveDataParam string) {
		test.T = t
		headers := make(map[string][]string, 0)
		if saveDataHeaderEnabled {
			headers["Save-Data"] = []string{"On"}
		}
		headers["Accept"] = []string{acceptFormats}

		urlStr := fmt.Sprintf("http://localhost/img/http%%3A%%2F%%2Fsite.com/img.png/resize?size=%s&dppx=%f&save-data=%s", size, dppx, saveDataParam)
		u, _ := url.Parse(urlStr)

		if u != nil {
			req := &http.Request{
				Method: "GET",
				URL:    u,
				Header: headers,
			}

			resp := httptest.NewRecorder()
			s(resp, req)

			statusCode := resp.Result().StatusCode
			switch {
			case statusCode >= http.StatusInternalServerError:
				t.Errorf("should not respond with 5xx errors, url: [%s] [%v]", urlStr, saveDataHeaderEnabled)
			case statusCode == http.StatusOK:
				contentType := resp.Header().Get("content-type")
				if contentType != "image/png" && contentType != "image/webp" && contentType != "image/avif" && contentType != "text/plain; charset=utf-8" {
					t.Errorf("unexpected Content-Type [%s] for URL [%s]", contentType, urlStr)
				}
			case statusCode == http.StatusBadRequest:
				t.Logf("Bad request for url: [%s]", urlStr)
			default:
				t.Errorf("unexpected response status [%d]", statusCode)
			}
		}
	})
}

func createService(t *testing.T) *img.Service {
	img.CacheTTL = 86400
	s, err := img.NewService(&loaderMock{}, &resizerMock{}, 1)
	if err != nil && t != nil {
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
