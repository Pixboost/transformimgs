package img

import (
	"bytes"
	"github.com/golang/glog"
	"log"
	"os/exec"
	"github.com/pkg/errors"
)

type ImageMagickProcessor struct {
	convertCmd string
}

var convertOpts = []string{
	"-filter", "Triangle",
	"-define", "filter:support=2",
	"-unsharp", "0.25x0.08+8.3+0.045",
	"-dither", "None",
	"-posterize", "136",
	"-quality", "82",
	"-define", "jpeg:fancy-upsampling=off",
	"-define", "png:compression-filter=5",
	"-define", "png:compression-level=9",
	"-define", "png:compression-strategy=1",
	"-define", "png:exclude-chunk=all",
	"-interlace", "Plane",
	"-colorspace", "sRGB",
	"-sampling-factor", "4:2:0",
	"-strip",
}

var cutToFitOpts = []string{
	"-gravity", "center",
}

//Creates new imagemagick processor. im is a path to
//ImageMagick executable that must be provided.
func NewProcessor(im string) (*ImageMagickProcessor, error) {
	if len(im) == 0 {
		log.Fatal("Command convert should be set by -imConvert flag")
		return nil, errors.New("Path to imagemagick executable must be provided")
	}

	_, err := exec.LookPath(im)
	if err != nil {
		return nil, err
	}

	return &ImageMagickProcessor{
		convertCmd: im,
	}, nil
}

// Resize image to the given size preserving aspect ratio. No cropping applying.
func (p *ImageMagickProcessor) Resize(data []byte, size string) ([]byte, error) {
	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size)
	args = append(args, convertOpts...)
	args = append(args, "-") //Output

	return p.execImagemagick(bytes.NewReader(data), args)
}

// Resize input image to exact size with cropping everything that out of the bounds.
// Size must specified in format WIDTHxHEIGHT. Both dimensions must be included.
func (p *ImageMagickProcessor) FitToSize(data []byte, size string) ([]byte, error) {
	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size+"^")
	args = append(args, convertOpts...)
	args = append(args, cutToFitOpts...)
	args = append(args, "-extent", size)
	args = append(args, "-") //Output

	return p.execImagemagick(bytes.NewReader(data), args)
}

func (p *ImageMagickProcessor) Optimise(data []byte) ([]byte, error) {
	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, convertOpts...)
	args = append(args, "-") //Output

	return p.execImagemagick(bytes.NewReader(data), args)
}

func (p *ImageMagickProcessor) execImagemagick(in *bytes.Reader, args []string) ([]byte, error) {
	var out, cmderr bytes.Buffer
	cmd := exec.Command(p.convertCmd)

	cmd.Args = append(cmd.Args, args...)

	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = &cmderr

	glog.Infof("Running resize command, args '%v'", cmd.Args)
	err := cmd.Run()
	if err != nil {
		glog.Errorf("Error executing convert command: %s", err.Error())
		glog.Errorf("ERROR: %s", cmderr.String())
		return nil, err
	}

	return out.Bytes(), nil
}
