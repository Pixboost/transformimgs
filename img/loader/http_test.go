package loader_test

import (
	"context"
	"github.com/Pixboost/transformimgs/img/loader"
	"github.com/dooman87/kolibri/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttp_LoadImg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "cool/stuff")
		w.Write([]byte("123"))
	}))
	defer server.Close()

	httpLoader := &loader.Http{}

	image, err := httpLoader.Load(server.URL, context.Background())

	test.Error(t,
		test.Nil(err, "error"),
		test.Equal("cool/stuff", image.MimeType, "content type"),
		test.Equal("123", string(image.Data), "resulted image"),
	)
}

func TestHttp_LoadImgErrorResponseStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	httpLoader := &loader.Http{}

	_, err := httpLoader.Load(server.URL, context.Background())

	test.Error(t,
		test.NotNil(err, "error"),
	)
}

func TestHttp_LoadCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("this-is-header") != "wow" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "cool/stuff")
			w.Write([]byte("123"))
		}
	}))
	defer server.Close()

	httpLoader := &loader.Http{
		Headers: http.Header{
			"This-Is-Header": []string{
				"wow",
			},
		},
	}

	image, err := httpLoader.Load(server.URL, context.Background())

	test.Error(t,
		test.Nil(err, "error"),
		test.Equal("cool/stuff", image.MimeType, "content type"),
		test.Equal("123", string(image.Data), "resulted image"),
	)
}
