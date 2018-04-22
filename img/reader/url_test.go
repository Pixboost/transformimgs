package reader_test

import (
	"github.com/Pixboost/transformimgs/img/reader"
	"github.com/dooman87/kolibri/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadImg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("123"))
	}))
	defer server.Close()

	urlReader := &reader.Http{}

	r, err := urlReader.Read(server.URL)

	test.Error(t,
		test.Nil(err, "error"),
		test.Equal("123", string(r), "resulted image"),
	)
}

func TestReadImgErrorResponseStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	urlReader := &reader.Http{}

	_, err := urlReader.Read(server.URL)

	test.Error(t,
		test.NotNil(err, "error"),
	)
}
