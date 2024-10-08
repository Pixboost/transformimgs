package processor_test

import (
	"bytes"
	"fmt"
	"github.com/Pixboost/transformimgs/v8/img"
	"github.com/Pixboost/transformimgs/v8/img/processor"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type testTransformation struct {
	file                   string
	expectedOutputMimeType string
}

type testIsIllustration struct {
	file           string
	isIllustration bool
}

type result struct {
	file     string
	origSize int
	optSize  int
}

type transform func(orig []byte, imgId string) (*img.Image, error)

var (
	proc         *processor.ImageMagick
	procWithArgs *processor.ImageMagick
)

func TestMain(m *testing.M) {
	var err error

	proc, err = processor.NewImageMagick(os.ExpandEnv("${IM_HOME}/convert"), os.ExpandEnv("${IM_HOME}/identify"))
	if err != nil {
		fmt.Printf("Error while creating image processor: %+v", err)
		os.Exit(1)
	}

	procWithArgs, err = processor.NewImageMagick(os.ExpandEnv("${IM_HOME}/convert"), os.ExpandEnv("${IM_HOME}/identify"))
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

func BenchmarkImageMagickProcessor_Optimise_Avif(b *testing.B) {
	benchmarkWithFormats(b, []string{"image/avif"})
}

func BenchmarkImageMagickProcessor_Optimise_Jxl(b *testing.B) {
	benchmarkWithFormats(b, []string{"image/jxl"})
}

func benchmarkWithFormats(b *testing.B, formats []string) {
	f := fmt.Sprintf("%s/%s", "./test_files/transformations", "medium-jpeg.jpg")

	orig, err := ioutil.ReadFile(f)
	if err != nil {
		b.Errorf("Can't read file %s: %+v", f, err)
	}
	processor.Debug = false

	for i := 0; i < b.N; i++ {
		_, err = proc.Optimise(&img.TransformationConfig{
			Src: &img.Image{
				Id:   f,
				Data: orig,
			},
			SupportedFormats: formats,
			Quality:          0,
			Config:           nil,
		})
		if err != nil {
			b.Errorf("Can't transform file: %+v", err)
		}
	}

	processor.Debug = true
}

func TestImageMagick_GetAdditionalArgs(t *testing.T) {
	var (
		aOp     string
		aImage  []byte
		aSource *img.Info
		aTarget *img.Info
	)
	proc.GetAdditionalArgs = func(op string, image []byte, source *img.Info, target *img.Info) []string {
		aOp = op
		aImage = image
		aSource = source
		aTarget = target
		return []string{}
	}

	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.Resize(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{},
			Config:           &img.ResizeConfig{Size: "50"},
		})
	}, []*testTransformation{{"opaque-png.png", ""}})

	if aOp != "resize" {
		t.Errorf("Expected op to be resize, but got [%s]", aOp)
	}
	if len(aImage) != 201318 {
		t.Errorf("Expected source image to be 201318 bytes, but got [%d]", len(aImage))
	}
	if !reflect.DeepEqual(aSource, &img.Info{
		Format:  "PNG",
		Quality: 100,
		Opaque:  true,
		Width:   400,
		Height:  400,
		Size:    201318,
	}) {
		t.Errorf("Source image error: %+v", aSource)
	}
	if !reflect.DeepEqual(aTarget, &img.Info{
		Format:  "",
		Quality: 0,
		Opaque:  true,
		Width:   50,
		Height:  50,
		Size:    0,
	}) {
		t.Errorf("Target image error: %+v", aTarget)
	}

	proc.GetAdditionalArgs = nil
}

func TestImageMagickProcessor_NoAccept(t *testing.T) {
	tests := []*testTransformation{
		{"big-jpeg.jpg", ""},
		{"opaque-png.png", ""},
		{"transparent-png-use-original.png", ""},
		{"small-transparent-png.png", ""},
		{"animated.gif", ""},
		{"animated-coalesce.gif", ""},
		{"logo.png", ""},
	}

	t.Run("optimise", func(t *testing.T) {
		testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
			return proc.Optimise(&img.TransformationConfig{
				Src: &img.Image{
					Id:   imgId,
					Data: orig,
				},
				SupportedFormats: []string{},
			})
		}, tests)

		testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
			return procWithArgs.Optimise(&img.TransformationConfig{
				Src: &img.Image{
					Id:   imgId,
					Data: orig,
				},
				SupportedFormats: []string{},
			})
		}, tests)
	})

	t.Run("resize", func(t *testing.T) {
		testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
			return proc.Resize(&img.TransformationConfig{
				Src: &img.Image{
					Id:   imgId,
					Data: orig,
				},
				SupportedFormats: []string{},
				Config:           &img.ResizeConfig{Size: "50"},
			})
		}, tests)

		testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
			return procWithArgs.Resize(&img.TransformationConfig{
				Src: &img.Image{
					Id:   imgId,
					Data: orig,
				},
				SupportedFormats: []string{},
				Config:           &img.ResizeConfig{Size: "50"},
			})
		}, tests)
	})

	t.Run("fit", func(t *testing.T) {
		testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
			return proc.FitToSize(&img.TransformationConfig{
				Src: &img.Image{
					Id:   imgId,
					Data: orig,
				},
				SupportedFormats: []string{},
				Config:           &img.ResizeConfig{Size: "50x50"},
			})
		}, tests)

		testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
			return procWithArgs.FitToSize(&img.TransformationConfig{
				Src: &img.Image{
					Id:   imgId,
					Data: orig,
				},
				SupportedFormats: []string{},
				Config:           &img.ResizeConfig{Size: "50x50"},
			})
		}, tests)
	})
}

func TestImageMagickProcessor_Optimise_Webp(t *testing.T) {
	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.Optimise(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{"image/webp"},
		})
	},
		[]*testTransformation{
			{"big-jpeg.jpg", "image/webp"},
			{"opaque-png.png", "image/webp"},
			{"transparent-png-use-original.png", "image/webp"},
			{"animated.gif", "image/webp"},
			{"animated-coalesce.gif", ""},
			{"webp-invalid-height.jpg", ""},
			{"logo.png", "image/webp"},
		})
}

func TestImageMagickProcessor_Resize_Webp(t *testing.T) {
	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.Resize(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{"image/webp"},
			Config:           &img.ResizeConfig{Size: "50"},
		})
	},
		[]*testTransformation{
			{"big-jpeg.jpg", "image/webp"},
			{"opaque-png.png", "image/webp"},
			{"transparent-png-use-original.png", "image/webp"},
			{"animated.gif", "image/webp"},
			{"animated-coalesce.gif", "image/webp"},
			{"webp-invalid-height.jpg", ""},
			{"logo.png", "image/webp"},
		})
}

func TestImageMagickProcessor_FitToSize_Webp(t *testing.T) {
	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.FitToSize(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{"image/webp"},
			Config:           &img.ResizeConfig{Size: "50x50"},
		})
	},
		[]*testTransformation{
			{"big-jpeg.jpg", "image/webp"},
			{"opaque-png.png", "image/webp"},
			{"transparent-png-use-original.png", "image/webp"},
			{"animated.gif", "image/webp"},
			{"animated-coalesce.gif", "image/webp"},
			{"webp-invalid-height.jpg", ""},
			{"logo.png", "image/webp"},
		})
}

func TestImageMagickProcessor_Optimise_Avif(t *testing.T) {
	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.Optimise(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{"image/avif"},
		})
	},
		[]*testTransformation{
			{"big-jpeg.jpg", ""},
			{"medium-jpeg.jpg", "image/avif"},
			{"opaque-png.png", "image/avif"},
			{"animated.gif", ""},
			{"animated-coalesce.gif", ""},
			{"transparent-png.png", "image/avif"},
			{"small-transparent-png.png", ""},
			{"transparent-png-use-original.png", ""},
			{"logo.png", ""},
		})
}

func TestImageMagickProcessor_Optimise_Jxl(t *testing.T) {
	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.Optimise(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{"image/jxl"},
		})
	},
		[]*testTransformation{
			{"big-jpeg.jpg", ""},
			{"medium-jpeg.jpg", "image/jxl"},
			{"opaque-png.png", "image/jxl"},
			{"animated.gif", ""},
			{"animated-coalesce.gif", ""},
			{"transparent-png.png", "image/jxl"},
			{"small-transparent-png.png", "image/jxl"},
			{"transparent-png-use-original.png", ""},
			{"logo.png", "image/jxl"},
		})
}

func TestImageMagickProcessor_Optimise_Avif_Webp(t *testing.T) {
	qualities := []img.Quality{img.DEFAULT, img.LOW, img.LOWER}

	for _, q := range qualities {
		t.Run(fmt.Sprintf("Quality_%d", q), func(t *testing.T) {
			testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
				return proc.Optimise(&img.TransformationConfig{
					Src: &img.Image{
						Id:   imgId,
						Data: orig,
					},
					Quality:          q,
					SupportedFormats: []string{"image/avif", "image/webp"},
				})
			},
				[]*testTransformation{
					{"big-jpeg.jpg", "image/webp"},
					{"medium-jpeg.jpg", "image/avif"},
					{"opaque-png.png", "image/avif"},
					{"animated.gif", "image/webp"},
					{"animated-coalesce.gif", "image/webp"},
					{"transparent-png.png", "image/avif"},
					{"small-transparent-png.png", "image/webp"},
					{"transparent-png-use-original.png", "image/webp"},
					{"logo.png", "image/webp"},
				})
		})
	}
}

func TestImageMagickProcessor_Optimise_Jxl_Avif_Webp(t *testing.T) {
	qualities := []img.Quality{img.DEFAULT, img.LOW, img.LOWER}

	for _, q := range qualities {
		t.Run(fmt.Sprintf("Quality_%d", q), func(t *testing.T) {
			testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
				return proc.Optimise(&img.TransformationConfig{
					Src: &img.Image{
						Id:   imgId,
						Data: orig,
					},
					Quality:          q,
					SupportedFormats: []string{"image/jxl", "image/avif", "image/webp"},
				})
			},
				[]*testTransformation{
					{"big-jpeg.jpg", "image/webp"},
					{"medium-jpeg.jpg", "image/avif"},
					{"opaque-png.png", "image/avif"},
					{"animated.gif", "image/webp"},
					{"animated-coalesce.gif", "image/webp"},
					{"transparent-png.png", "image/avif"},
					{"small-transparent-png.png", "image/jxl"},
					{"transparent-png-use-original.png", ""},
					{"logo.png", "image/jxl"},
				})
		})
	}
}

func TestImageMagickProcessor_Optimise_Jxl_Webp(t *testing.T) {
	qualities := []img.Quality{img.DEFAULT, img.LOW, img.LOWER}

	for _, q := range qualities {
		t.Run(fmt.Sprintf("Quality_%d", q), func(t *testing.T) {
			testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
				return proc.Optimise(&img.TransformationConfig{
					Src: &img.Image{
						Id:   imgId,
						Data: orig,
					},
					Quality:          q,
					SupportedFormats: []string{"image/jxl", "image/webp"},
				})
			},
				[]*testTransformation{
					{"big-jpeg.jpg", "image/webp"},
					{"medium-jpeg.jpg", "image/jxl"},
					{"opaque-png.png", "image/jxl"},
					{"animated.gif", "image/webp"},
					{"animated-coalesce.gif", "image/webp"},
					{"transparent-png.png", "image/jxl"},
					{"small-transparent-png.png", "image/jxl"},
					{"transparent-png-use-original.png", ""},
					{"logo.png", "image/jxl"},
				})
		})
	}
}

func TestImageMagickProcessor_Resize_Avif(t *testing.T) {
	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.Resize(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{"image/avif"},
			Config:           &img.ResizeConfig{Size: "50"},
		})
	},
		[]*testTransformation{
			{"big-jpeg.jpg", "image/avif"},
			{"medium-jpeg.jpg", "image/avif"},
			{"opaque-png.png", "image/avif"},
			{"animated.gif", ""},
			{"animated-coalesce.gif", ""},
			{"transparent-png-use-original.png", ""},
			{"logo.png", ""},
		})
}

func TestImageMagickProcessor_Resize_Jxl(t *testing.T) {
	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.Resize(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{"image/jxl"},
			Config:           &img.ResizeConfig{Size: "50"},
		})
	},
		[]*testTransformation{
			{"big-jpeg.jpg", "image/jxl"},
			{"medium-jpeg.jpg", "image/jxl"},
			{"opaque-png.png", "image/jxl"},
			{"animated.gif", ""},
			{"transparent-png-use-original.png", "image/jxl"},
			{"logo.png", "image/jxl"},
		})
}

func TestImageMagickProcessor_FitToSize_Jxl(t *testing.T) {
	testImages(t, func(orig []byte, imgId string) (*img.Image, error) {
		return proc.FitToSize(&img.TransformationConfig{
			Src: &img.Image{
				Id:   imgId,
				Data: orig,
			},
			SupportedFormats: []string{"image/jxl"},
			Config:           &img.ResizeConfig{Size: "50x50"},
		})
	},
		[]*testTransformation{
			{"big-jpeg.jpg", "image/jxl"},
			{"medium-jpeg.jpg", "image/jxl"},
			{"opaque-png.png", "image/jxl"},
			{"animated.gif", ""},
			{"animated-coalesce.gif", ""},
			{"transparent-png-use-original.png", "image/jxl"},
			{"logo.png", "image/jxl"},
		})
}

var isIllustrationTests = []*testIsIllustration{
	{"illustration-1.png", true},
	{"illustration-2.png", true},
	{"illustration-3.png", true},
	{"logo-1.png", true},
	{"logo-2.png", true},
	{"banner-1.png", false},
	{"screenshot-1.png", false},
	{"photo-1.png", false},
	{"photo-2.png", false},
	{"photo-3.png", false},
	{"product-1.png", false},
	{"product-2.png", false},
	{"product-2-no-background.png", false},
	{"product-3.png", false},
}

func TestImageMagick_IsIllustration(t *testing.T) {
	for _, tt := range isIllustrationTests {
		imgFile := tt.file

		f := fmt.Sprintf("%s/%s", "./test_files/is_illustration", imgFile)

		orig, err := ioutil.ReadFile(f)
		if err != nil {
			t.Errorf("Can't read file %s: %+v", f, err)
		}

		image := &img.Image{
			Id:       imgFile,
			Data:     orig,
			MimeType: "",
		}
		info, err := proc.LoadImageInfo(image)
		if err != nil {
			t.Errorf("could not load image info %s: %s", imgFile, err)
		}

		if err != nil {
			t.Errorf("Unexpected error [%s]: %s", imgFile, err)
		}
		if info.Illustration != tt.isIllustration {
			t.Errorf("Expected [%t] for [%s], but got [%t]", tt.isIllustration, imgFile, info.Illustration)
		}
	}
}

var trimBorderTestFiles = []string{"logo-1.png", "logo-2.png", "no-border.jpg"}

func TestImageMagick_TrimBorder(t *testing.T) {
	var overrideExpected = false

	for _, tt := range trimBorderTestFiles {
		f := fmt.Sprintf("%s/%s", "./test_files/trim-border", tt)

		orig, err := ioutil.ReadFile(f)
		if err != nil {
			t.Errorf("can't read file %s: %+v", f, err)
		}

		resultImage, err := proc.Optimise(&img.TransformationConfig{
			Src: &img.Image{
				Id:   "img",
				Data: orig,
			},
			TrimBorder: true,
		})

		if err != nil {
			t.Errorf("couldn't optimise image %s", tt)
		}

		expectedFile := fmt.Sprintf("./test_files/trim-border/expected/optimise/%s", tt)
		compareImage(resultImage, expectedFile, t, overrideExpected)

		resultImage, err = proc.Resize(&img.TransformationConfig{
			Src: &img.Image{
				Id:   "img",
				Data: orig,
			},
			TrimBorder: true,
			Config:     &img.ResizeConfig{Size: "300"},
		})

		if err != nil {
			t.Errorf("couldn't resize image %s", tt)
		}

		expectedFile = fmt.Sprintf("./test_files/trim-border/expected/resize/%s", tt)
		compareImage(resultImage, expectedFile, t, overrideExpected)

		resultImage, err = proc.FitToSize(&img.TransformationConfig{
			Src: &img.Image{
				Id:   "img",
				Data: orig,
			},
			TrimBorder: true,
			Config:     &img.ResizeConfig{Size: "200x80"},
		})

		if err != nil {
			t.Errorf("couldn't fit image %s", tt)
		}

		expectedFile = fmt.Sprintf("./test_files/trim-border/expected/fit/%s", tt)
		compareImage(resultImage, expectedFile, t, overrideExpected)
	}
}

func compareImage(img *img.Image, expectedFile string, t *testing.T, overrideExpected bool) {
	if overrideExpected {
		ioutil.WriteFile(expectedFile, img.Data, 0777)
	} else {
		ext := filepath.Ext(expectedFile)
		actualFile, err := os.CreateTemp("", fmt.Sprintf("image*%s", ext))
		if err != nil {
			t.Errorf("could not create temp file %s", err)
		}
		_, err = actualFile.Write(img.Data)
		if err != nil {
			t.Errorf("could not write to temp file %s", err)
		}
		_ = actualFile.Close()

		var out, cmderr bytes.Buffer
		cmd := exec.Command(os.ExpandEnv("${IM_HOME}/magick"))
		cmd.Args = append(cmd.Args, "compare", "-metric", "AE", actualFile.Name(), expectedFile, "null:")
		cmd.Stdout = &out
		cmd.Stderr = &cmderr

		fmt.Println(cmd.Args)

		err = cmd.Run()
		errStr := cmderr.String()
		outStr := out.String()
		if err != nil {
			t.Errorf("error executing compare command: %s, %s", err.Error(), errStr)
		}
		pixelsDiffCnt, err := strconv.Atoi(errStr)
		if err != nil {
			t.Errorf("could not parse output of compare [%s]", outStr)
		}
		if pixelsDiffCnt > 0 {
			t.Errorf("expected 0 different pixels but found %d when comparing %s", pixelsDiffCnt, expectedFile)
		}
	}
}

func testImages(t *testing.T, fn transform, files []*testTransformation) {
	results := make([]*result, 0)
	for _, tt := range files {
		imgFile := tt.file

		f := fmt.Sprintf("%s/%s", "./test_files/transformations", imgFile)

		orig, err := ioutil.ReadFile(f)
		if err != nil {
			t.Errorf("Can't read file %s: %+v", f, err)
		}

		transformedImg, err := fn(orig, f)

		if err != nil {
			t.Errorf("Can't transform file: %+v", err)
		}

		results = append(results, &result{
			file:     imgFile,
			origSize: len(orig),
			optSize:  len(transformedImg.Data),
		})
		//Writes converted file for manual verification.
		// ioutil.WriteFile(fmt.Sprintf("./test_files/opt_%s_%s", t.Name(), imgFile), transformedImg, 0777)

		if transformedImg.MimeType != tt.expectedOutputMimeType {
			t.Errorf("%s: Expected [%s] mime type, but got [%s]", tt.file, tt.expectedOutputMimeType, transformedImg.MimeType)
		}

		if len(transformedImg.Data) > len(orig) {
			t.Errorf("Image %s is not optimised", f)
		}
	}

	for _, r := range results {
		fmt.Printf("%60s | %10d | %10d | %.2f\n", r.file, r.optSize, r.origSize, 1.0-(float32(r.optSize)/float32(r.origSize)))
	}
}

func TestOptimise_ErrorIdentify(t *testing.T) {
	_, err := proc.Optimise(&img.TransformationConfig{
		Src: &img.Image{
			Id:       "",
			Data:     []byte("This is not an image!"),
			MimeType: "image/jpeg",
		},
	})

	if err == nil {
		t.Error("expected error but got none")
	}

	expectedError := "identify: no decode delegate for this image format"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error to contain [%s], but got [%s]", expectedError, err.Error())
	}
}

func TestOptimise_ErrorConvert(t *testing.T) {
	f := fmt.Sprintf("%s/%s", "./test_files/transformations", "medium-jpeg.jpg")

	orig, err := ioutil.ReadFile(f)
	if err != nil {
		t.Errorf("Can't read file %s: %+v", f, err)
	}

	proc.GetAdditionalArgs = func(op string, image []byte, source *img.Info, target *img.Info) []string {
		return []string{"-i_dont_know", "this"}
	}

	_, err = proc.Resize(&img.TransformationConfig{
		Src: &img.Image{
			Id:       "",
			Data:     orig,
			MimeType: "image/jpeg",
		},
		Config: &img.ResizeConfig{
			Size: "300x300",
		},
	})

	proc.GetAdditionalArgs = nil

	if err == nil {
		t.Error("expected error but got none")
	}

	expectedError := "unrecognized option `-i_dont_know'"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error to contain [%s], but got [%s]", expectedError, err.Error())
	}
}
