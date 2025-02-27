package main

import (
	"flag"
	"github.com/Pixboost/transformimgs/v8/img"
	"github.com/Pixboost/transformimgs/v8/img/loader"
	"github.com/Pixboost/transformimgs/v8/img/processor"
	"github.com/dooman87/kolibri/health"
	"net/http"
	"os"
	"runtime"
	"time"
)

func main() {
	var (
		im              string
		imIdent         string
		cache           int
		procNum         int
		disableSaveData bool
	)
	flag.StringVar(&im, "imConvert", "", "Imagemagick convert command")
	flag.StringVar(&imIdent, "imIdentify", "", "Imagemagick identify command")
	flag.IntVar(&cache, "cache", 2592000,
		"Number of seconds to cache image after transformation (0 to disable cache). Default value is 2592000 (30 days)")
	flag.IntVar(&procNum, "proc", runtime.NumCPU(), "Number of images processors to run. Defaults to number of CPUs")
	flag.BoolVar(&disableSaveData, "disableSaveData", false, "If set to true then will disable Save-Data client hint. Could be useful for CDNs that don't support Save-Data header in Vary.")
	flag.Parse()

	p, err := processor.NewImageMagick(im, imIdent)

	if err != nil {
		img.Log.Errorf("Can't create image magic processor: %+v", err)
		os.Exit(1)
	}

	img.CacheTTL = cache
	img.SaveDataEnabled = !disableSaveData
	srv, err := img.NewService(&loader.Http{}, p, procNum)
	if err != nil {
		img.Log.Errorf("Can't create image service: %+v", err)
		os.Exit(2)
	}

	router := srv.GetRouter()
	router.HandleFunc("/health", health.Health)

	img.Log.Printf("Running the application on port 8080...\n")
	server := http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err = server.ListenAndServe()

	if err != nil {
		img.Log.Errorf("Error while stopping application: %+v", err)
		os.Exit(3)
	}
	os.Exit(0)
}
