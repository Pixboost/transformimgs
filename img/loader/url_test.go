package loader_test

import (
	"github.com/Pixboost/transformimgs/img/loader"
	"github.com/dooman87/kolibri/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadImg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "cool/stuff")
		w.Write([]byte("123"))
	}))
	defer server.Close()

	urlReader := &loader.Http{}

	r, contentType, err := urlReader.Read(server.URL)

	test.Error(t,
		test.Nil(err, "error"),
		test.Equal("cool/stuff", contentType, "content type"),
		test.Equal("123", string(r), "resulted image"),
	)
}

func TestReadImgErrorResponseStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	urlReader := &loader.Http{}

	_, _, err := urlReader.Read(server.URL)

	test.Error(t,
		test.NotNil(err, "error"),
	)
}

func TestCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("this-is-header") != "wow" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "cool/stuff")
			w.Write([]byte("123"))
		}
	}))
	defer server.Close()

	urlReader := &loader.Http{
		Headers: http.Header{
			"This-Is-Header": []string{
				"wow",
			},
		},
	}

	r, contentType, err := urlReader.Read(server.URL)

	test.Error(t,
		test.Nil(err, "error"),
		test.Equal("cool/stuff", contentType, "content type"),
		test.Equal("123", string(r), "resulted image"),
	)
}
