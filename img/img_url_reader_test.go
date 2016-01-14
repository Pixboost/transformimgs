package img_test

import (
	"github.com/dooman87/transformimgs/img"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadImg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("123"))
	}))
	defer server.Close()

	reader := &img.ImgUrlReader{}

	img, err := reader.Read(server.URL)

	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := "123"
	actual := string(img)
	if expected != actual {
		t.Fatalf("Expected %s but got %s", expected, actual)
	}
}

func TestReadImgErrorResponseStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	reader := &img.ImgUrlReader{}

	_, err := reader.Read(server.URL)

	if err == nil {
		t.Fatalf("Expected error")
	}
}
