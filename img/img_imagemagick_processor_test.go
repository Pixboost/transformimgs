package img_test

import (
	"fmt"
	"github.com/dooman87/transformimgs/img"
	"io/ioutil"
	"log"
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

func TestImageMagickProcessor_Optimise(t *testing.T) {
	proc, err := img.NewProcessor("c:/Program Files/ImageMagick-7.0.5-Q16/convert", "c:/Program Files/ImageMagick-7.0.5-Q16/identify")

	if err != nil {
		t.Errorf("Error while creating image processor: %+v", err)
	}

	results := make([]*result, 0)
	for _, imgFile := range FILES {
		f := fmt.Sprintf("%s/%s", "./test_files", imgFile)
		optimisedImg, orig, err := opt(proc, f)
		if err != nil {
			t.Errorf("Can't optimise file: %+v", err)
		}

		results = append(results, &result{
			file:     imgFile,
			origSize: len(orig),
			optSize:  len(optimisedImg),
		})
		//log.Printf("Optimised size: %d, original: %d", len(optimisedImg), len(img))
		//ioutil.WriteFile(fmt.Sprintf("./test_files/opt_%s", imgFile), optimisedImg, 0777)

		if len(optimisedImg) > len(orig) {
			t.Errorf("Image %s is not optimised", f)
		}
	}

	for _, r := range results {
		log.Printf("%60s | %10d | %10d | %.2f", r.file, r.optSize, r.origSize, 1.0-(float32(r.optSize)/float32(r.origSize)))
	}
}

func opt(proc img.ImgProcessor, file string) ([]byte, []byte, error) {
	orig, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, orig, err
	}

	optimisedImg, err := proc.Optimise(orig)

	return optimisedImg, orig, err
}
