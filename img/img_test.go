package img_test

import (
	"errors"
	"github.com/dooman87/kolibri/test"
	"github.com/dooman87/transformimgs/img"
	"net/http"
	"testing"
	"net/http/httptest"
)

type resizerMock struct{}

func (r *resizerMock) Resize(data []byte, size string) ([]byte, error) {
	if string(data) == "321" && size == "300x200" {
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

func TestResize(t *testing.T) {
	service := &img.Service{
		Reader:    &readerMock{},
		Processor: &resizerMock{},
	}
	test.Service = service.ResizeUrl

	testCases := []test.TestCase{
		{"http://localhost/img?url=http://site.com/img.png&size=300x200", http.StatusOK, "Success",
			func(w *httptest.ResponseRecorder, t *testing.T) {
				if w.Header().Get("Cache-Control") != "max-age=86400" {
					t.Errorf("Expected to get Cache-Control header")
				}
				if w.Header().Get("Content-Length") != "3" {
					t.Errorf("Expected to get Content-Length header equal to 3 but got [%s]", w.Header().Get("Content-Length"))
				}
			}},
		{"http://localhost/img?size=300x200", http.StatusBadRequest, "Param url is required", nil},
		{"http://localhost/img?url=http://site.com/img.png", http.StatusBadRequest, "Param size is required", nil},
		{"http://localhost/img?url=NO_SUCH_IMAGE&size=300x200", http.StatusInternalServerError, "Read error", nil},
		{"http://localhost/img?url=http://site.com/img.png&size=BADSIZE", http.StatusInternalServerError, "Resize error", nil},
	}

	test.RunRequests(testCases, t)
}
