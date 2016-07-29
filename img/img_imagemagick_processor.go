package img

import (
	"bytes"
	"flag"
	"github.com/golang/glog"
	"log"
	"os/exec"
)

type ImageMagickProcessor struct {
	convertCmd string
}

var imagemagickConvertCmd string
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

func init() {
	flag.StringVar(&imagemagickConvertCmd, "imConvert", "", "Imagemagick convert command")
}

//Checks that image magick is available.
// If it's not then terminating application with fatal logging.
func CheckImagemagick() {
	if len(imagemagickConvertCmd) == 0 {
		log.Fatal("Command convert should be set by -imConvert flag")
		return
	}

	_, err := exec.LookPath(imagemagickConvertCmd)
	if err != nil {
		log.Fatalf("Imagemagick is not available '%s'", err.Error())
	}
}

// Resize image to the given size preserving aspect ratio. No cropping applying.
func (p *ImageMagickProcessor) Resize(data []byte, size string) ([]byte, error) {
	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size)
	args = append(args, convertOpts...)
	args = append(args, "-") //Output

	return execImagemagick(bytes.NewReader(data), args)
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

	return execImagemagick(bytes.NewReader(data), args)
}

func (p *ImageMagickProcessor) Optimise(data []byte) ([]byte, error) {
	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, convertOpts...)
	args = append(args, "-") //Output

	return execImagemagick(bytes.NewReader(data), args)
}

func execImagemagick(in *bytes.Reader, args []string) ([]byte, error) {
	var out, cmderr bytes.Buffer
	cmd := exec.Command(imagemagickConvertCmd)

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
