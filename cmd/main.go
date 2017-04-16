// Images transformation API
//
// The purpose of this API is to provide a set of
// endpoints that will transform and optimise images.
// Then it's become easy to use the API with <img> tag.
// It's can be accessed directly through HTTP interface as
// well.
//
// Version: 1
// Schemes: http
// Host: pixboost.com
// BasePath: /api/1/
// Security:
// - api_key:
// SecurityDefinitions:
// - api_key:
//   type: apiKey
//   name: auth
//   in: query
// swagger:meta
package main

import (
	"flag"
	"github.com/dooman87/kolibri/health"
	"github.com/dooman87/transformimgs/img"
	"log"
	"net/http"
)

func main() {
	var (
		im    string
		cache int
	)
	flag.StringVar(&im, "imConvert", "", "Imagemagick convert command")
	flag.IntVar(&cache, "cache", 86400,
		"Number of seconds to cache image after transformation (0 to disable cache). Default value is 86400 (one day)")
	flag.Parse()

	p, err := img.NewProcessor(im)
	if err != nil {
		log.Fatalf("Can't create image magic processor: %+v", err)
	}

	srv, err := img.NewService(&img.ImgUrlReader{}, p, cache)
	if err != nil {
		log.Fatalf("Can't create image service: %+v", err)
	}

	http.HandleFunc("/health", health.Health)
	http.Handle("/", srv.GetRouter())

	log.Println("Running the applicaiton on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
