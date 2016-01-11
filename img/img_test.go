package img_test
import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/dooman87/transformimgs/img"
	"errors"
)

type resizerMock struct {}
func (r *resizerMock) Resize(data []byte, width int, height int) ([]byte, error) {
	if string(data) == "321" && width == 300 && height == 200 {
		return []byte("123"), nil
	}
	return nil, errors.New("error")
}

type readerMock struct {}
func (r *readerMock) Read(url string) ([]byte, error) {
	if url == "http://site.com/img.png" {
		return []byte("321"), nil
	}
	return nil, errors.New("error")
}

func TestResize(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost/img?url=http://site.com/img.png&width=300&height=200", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	service := &img.Service {
		Reader: &readerMock{},
		Processor: &resizerMock{},
	}
	service.ResizeUrl(w, req)

	expected := "123"
	if w.Body.String() != expected {
		t.Fatalf("Expected %s but got %s", expected, w.Body.String())
	}
}

func TestResizeNoUrl(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost/img?width=300&height=200", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	service := &img.Service {
		Reader: &readerMock{},
		Processor: &resizerMock{},
	}
	service.ResizeUrl(w, req)

	expected := http.StatusBadRequest
	if w.Code != expected {
		t.Fatalf("Expected %d but got %d", expected, w.Code)
	}
	expectedBody := "url param is required"
	if w.Body.String() != expectedBody {
		t.Fatalf("Expected '%s' but got '%s'", expectedBody, w.Body.String())
	}
}
