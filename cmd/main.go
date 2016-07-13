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
	flag.Parse()

	img.CheckImagemagick()

	imgService := img.Service{
		Processor: &img.ImageMagickProcessor{},
		Reader:    &img.ImgUrlReader{},
	}

	http.HandleFunc("/health", health.Health)
	http.Handle("/", imgService.GetRouter())

	glog.Info("Running the applicaiton on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
