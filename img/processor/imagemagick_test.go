package processor_test

import (
	"fmt"
	"github.com/Pixboost/transformimgs/img/processor"
	"io/ioutil"
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
		"Monochrome_CategoryImage2.png",
		"otto-brights-stationery.jpg",
		"ollie.png",
		"webp-invalid-height.jpg",
	}
)

type result struct {
	file     string
	origSize int
	optSize  int
}

type transform func(orig []byte, imgId string) ([]byte, error)

var (
	proc         *processor.ImageMagick
	procWithArgs *processor.ImageMagick
)

func TestMain(m *testing.M) {
	var err error

	proc, err = processor.NewImageMagick("/usr/bin/convert", "/usr/bin/identify")
	if err != nil {
		fmt.Printf("Error while creating image processor: %+v", err)
		os.Exit(1)
	}

	procWithArgs, err = processor.NewImageMagick("/usr/bin/convert", "/usr/bin/identify")
	if err != nil {
		fmt.Printf("Error while creating image processor: %+v", err)
		os.Exit(2)
	}
	procWithArgs.AdditionalArgs = []string{
		"-limit", "memory", "64MiB",
		"-limit", "map", "128MiB",
	}
	os.Exit(m.Run())
}

func BenchmarkImageMagickProcessor_Optimise(b *testing.B) {
	benchmarkWithFormats(b, []string{})
}

func BenchmarkImageMagickProcessor_Optimise_Webp(b *testing.B) {
	benchmarkWithFormats(b, []string{"image/webp"})
}

func benchmarkWithFormats(b *testing.B, formats []string) {
	f := fmt.Sprintf("%s/%s", "./test_files", "OW20170515_HPHB_B2B2.jpg")

	orig, err := ioutil.ReadFile(f)
	if err != nil {
		b.Errorf("Can't read file %s: %+v", f, err)
	}
	processor.Debug = false

	for i := 0; i < b.N; i++ {
		_, err = proc.Optimise(orig, f, formats)
		if err != nil {
			b.Errorf("Can't transform file: %+v", err)
		}
	}

	processor.Debug = true
}

func TestImageMagickProcessor_Optimise(t *testing.T) {
	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return proc.Optimise(orig, imgId, []string{})
	})

	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return procWithArgs.Optimise(orig, imgId, []string{})
	})
}

func TestImageMagickProcessor_Resize(t *testing.T) {
	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return proc.Resize(orig, "50", imgId, []string{})
	})

	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return procWithArgs.Resize(orig, "50", imgId, []string{})
	})
}

func TestImageMagickProcessor_FitToSize(t *testing.T) {
	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return proc.FitToSize(orig, "50x50", imgId, []string{})
	})

	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return procWithArgs.FitToSize(orig, "50x50", imgId, []string{})
	})
}

func TestImageMagickProcessor_Optimise_Webp(t *testing.T) {
	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return proc.Optimise(orig, imgId, []string{"image/webp"})
	})
}

func TestImageMagickProcessor_Resize_Webp(t *testing.T) {
	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return proc.Resize(orig, "50", imgId, []string{"image/webp"})
	})
}

func TestImageMagickProcessor_FitToSize_Webp(t *testing.T) {
	imgOpT(t, func(orig []byte, imgId string) ([]byte, error) {
		return proc.FitToSize(orig, "50x50", imgId, []string{"image/webp"})
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

		optimisedImg, err := fn(orig, f)

		if err != nil {
			t.Errorf("Can't transform file: %+v", err)
		}

		results = append(results, &result{
			file:     imgFile,
			origSize: len(orig),
			optSize:  len(optimisedImg),
		})
		//Writes converted file for manual verification.
		// ioutil.WriteFile(fmt.Sprintf("./test_files/opt_%s_%s", t.Name(), imgFile), optimisedImg, 0777)

		if len(optimisedImg) > len(orig) {
			t.Errorf("Image %s is not optimised", f)
		}
	}

	for _, r := range results {
		fmt.Printf("%60s | %10d | %10d | %.2f\n", r.file, r.optSize, r.origSize, 1.0-(float32(r.optSize)/float32(r.origSize)))
	}
}