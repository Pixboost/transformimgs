package img_test

import (
	"errors"
	"github.com/dooman87/transformimgs/img"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	test(t,
		"http://localhost/img?url=http://site.com/img.png&size=300x200",
		"123",
		http.StatusOK)
}

func TestResizeNoUrl(t *testing.T) {
	test(t,
		"http://localhost/img?size=300x200",
		"url param is required",
		http.StatusBadRequest)
}

func TestResizeNoSize(t *testing.T) {
	test(t,
		"http://localhost/img?url=http://site.com/img.png",
		"size param is required",
		http.StatusBadRequest)
}

func TestResizeReadError(t *testing.T) {
	test(t,
		"http://localhost/img?url=NO_SUCH_IMAGE&size=300x200",
		"Error reading image: 'read_error'",
		http.StatusInternalServerError)
}

func TestResizeResizeError(t *testing.T) {
	test(t,
		"http://localhost/img?url=http://site.com/img.png&size=BADSIZE",
		"Error transforming image: 'resize_error'",
		http.StatusInternalServerError)
}

func test(t *testing.T, url string, expectedResp string, expectedCode int) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	service := &img.Service{
		Reader:    &readerMock{},
		Processor: &resizerMock{},
	}
	service.ResizeUrl(w, req)

	if w.Code != expectedCode {
		t.Fatalf("Expected %d but got %d", expectedCode, w.Code)
	}

	if strings.Trim(w.Body.String(), " \n\r") != expectedResp {
		t.Fatalf("Expected '%s' but got '%s'", expectedResp, w.Body.String())
	}
}
