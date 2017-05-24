package img_test

import (
	"flag"
	"fmt"
	"github.com/dooman87/transformimgs/img"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var (
	FILES = []string{
		"HT_Paper.png",
		"HT_Stationery.png",
		"JBBAKUMBBK_baku_medium_back_chair_black.jpg",
		"otto-funhouse.jpg",
		"OW20170515_HPHB_B2B2.jpg",
		"OW20170515_HPHB_B2C4.jpg",
	}
)

type result struct {
	file     string
	origSize int
	optSize  int
}

type transform func(orig []byte) ([]byte, error)

var (
	proc *img.ImageMagickProcessor
)

func TestMain(m *testing.M) {
	var (
		err     error
		im      string
		imIdent string
	)

	flag.StringVar(&im, "imConvert", "", "Imagemagick convert command")
	flag.StringVar(&imIdent, "imIdentify", "", "Imagemagick identify command")
	flag.Parse()

	proc, err = img.NewProcessor(im, imIdent)
	if err != nil {
		log.Printf("Error while creating image processor: %+v", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestImageMagickProcessor_Optimise(t *testing.T) {
	imgOpT(t, func(orig []byte) ([]byte, error) {
		return proc.Optimise(orig)
	})
}

func BenchmarkImageMagickProcessor_Optimise(b *testing.B) {
	f := fmt.Sprintf("%s/%s", "./test_files", "OW20170515_HPHB_B2B2.jpg")

	orig, err := ioutil.ReadFile(f)
	if err != nil {
		b.Errorf("Can't read file %s: %+v", f, err)
	}
	img.Debug = false

	for i := 0; i < b.N; i++ {
		_, err = proc.Optimise(orig)
		if err != nil {
			b.Errorf("Can't transform file: %+v", err)
		}
	}

	img.Debug = true
}

func TestImageMagickProcessor_Resize(t *testing.T) {
	imgOpT(t, func(orig []byte) ([]byte, error) {
		return proc.Resize(orig, "50")
	})
}

func TestImageMagickProcessor_FitToSize(t *testing.T) {
	imgOpT(t, func(orig []byte) ([]byte, error) {
		return proc.FitToSize(orig, "50x50")
	})
}

func imgOpT(t *testing.T, fn transform) {
	results := make([]*result, 0)
	for _, imgFile := range FILES {
		f := fmt.Sprintf("%s/%s", "./test_files", imgFile)

		orig, err := ioutil.ReadFile(f)
		if err != nil {
			t.Errorf("Can't read file %s: %+v", f, err)
		}

		optimisedImg, err := fn(orig)

		if err != nil {
			t.Errorf("Can't transform file: %+v", err)
		}

		results = append(results, &result{
			file:     imgFile,
			origSize: len(orig),
			optSize:  len(optimisedImg),
		})
		//ioutil.WriteFile(fmt.Sprintf("./test_files/opt_%s", imgFile), optimisedImg, 0777)

		if len(optimisedImg) > len(orig) {
			t.Errorf("Image %s is not optimised", f)
		}
	}

	for _, r := range results {
		log.Printf("%60s | %10d | %10d | %.2f", r.file, r.optSize, r.origSize, 1.0-(float32(r.optSize)/float32(r.origSize)))
	}
}
