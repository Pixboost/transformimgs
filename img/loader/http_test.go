package loader_test

import (
	"context"
	"github.com/Pixboost/transformimgs/v8/img/loader"
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

func FuzzHttp_LoadImg(f *testing.F) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "cool/stuff")
		w.Write([]byte("123"))
	}))
	defer server.Close()

	httpLoader := &loader.Http{}

	f.Add("")
	f.Add("path/to/image.png")
	f.Add("//image.png,?&")
	f.Add("?!#")
	f.Fuzz(func(t *testing.T, path string) {
		image, err := httpLoader.Load(server.URL+"/"+path, nil)

		if err != nil {
			if image != nil {
				t.Errorf("image must be nil when error is not nil")
			}
		} else {
			test.Error(t,
				test.Nil(err, "error"),
				test.Equal("cool/stuff", image.MimeType, "content type"),
				test.Equal("123", string(image.Data), "resulted image"),
			)
		}
	})
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

func FuzzHttp_LoadCustomHeaders(f *testing.F) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "cool/stuff")
		w.Write([]byte("123"))
	}))
	defer server.Close()

	f.Add("custom-header", "value", uint(1))
	f.Add("custom-header-many-times", "value", uint(3))
	f.Fuzz(func(t *testing.T, name string, value string, count uint) {
		values := make([]string, count)
		for i := uint(0); i < count; i++ {
			values[i] = value
		}

		httpLoader := &loader.Http{
			Headers: http.Header{
				name: values,
			},
		}

		image, err := httpLoader.Load(server.URL, context.Background())

		if err != nil {
			if image != nil {
				t.Errorf("image must be nil when error is not nil")
			}
		} else {
			test.Error(t,
				test.Nil(err, "error"),
				test.Equal("cool/stuff", image.MimeType, "content type"),
				test.Equal("123", string(image.Data), "resulted image"),
			)
		}
	})
}
