package img_test

import (
	"errors"
	"github.com/dooman87/kolibri/test"
	"github.com/dooman87/transformimgs/img"
	"net/http"
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
	service := &img.Service{
		Reader:    &readerMock{},
		Processor: &resizerMock{},
	}
	test.Service = service.ResizeUrl

	testCases := []test.TestCase{
		{"http://localhost/img?url=http://site.com/img.png&size=300x200", http.StatusOK, "Success"},
		{"http://localhost/img?size=300x200", http.StatusBadRequest, "Param url is required"},
		{"http://localhost/img?url=http://site.com/img.png", http.StatusBadRequest, "Param size is required"},
		{"http://localhost/img?url=NO_SUCH_IMAGE&size=300x200", http.StatusInternalServerError, "Read error"},
		{"http://localhost/img?url=http://site.com/img.png&size=BADSIZE", http.StatusInternalServerError, "Resize error"},
	}

	test.RunRequests(testCases, t)
}
