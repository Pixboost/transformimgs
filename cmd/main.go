package main

import (
	"flag"
	"github.com/dooman87/kolibri/health"
	"github.com/dooman87/transformimgs/img"
	"github.com/golang/glog"
	"log"
	"net/http"
)

func main() {
	var (
		im string
		cache int
	)
	flag.StringVar(&im, "imConvert", "", "Imagemagick convert command")
	flag.IntVar(&cache, "cache", 86400,
		"Number of seconds to cache image after transformation (0 to disable cache). Default value is 86400 (one day)")
	flag.Parse()

	p, err := img.NewProcessor(im)
	if err != nil {
		glog.Fatalf("Can't create image magic processor: %+v", err)
	}

	srv, err := img.NewService(&img.ImgUrlReader{}, p, cache)
	if err != nil {
		glog.Fatalf("Can't create image service: %+v", err)
	}

	http.HandleFunc("/health", health.Health)
	http.Handle("/", srv.GetRouter())

	glog.Info("Running the applicaiton on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
