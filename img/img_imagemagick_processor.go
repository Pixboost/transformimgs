package img

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os/exec"
	"strconv"
)

type ImageMagickProcessor struct {
	convertCmd  string
	identifyCmd string
}

type imageInfo struct {
	format  string
	quality int
}

var convertOpts = []string{
	"-unsharp", "0.25x0.08+8.3+0.045",
	"-dither", "None",
	"-colors", "256",
	"-posterize", "136",
	"-define", "jpeg:fancy-upsampling=off",
	"-define", "png:compression-filter=5",
	"-define", "png:compression-level=9",
	"-define", "png:compression-strategy=0",
	"-interlace", "None",
	"-colorspace", "sRGB",
	"-sampling-factor", "4:2:0",
	"-strip",
	"+profile", "*",
}

var cutToFitOpts = []string{
	"-gravity", "center",
}

//If set then will print all commands to stdout.
var Debug bool = true

//Creates new imagemagick processor. im is a path to
//IM convert executable that must be provided.
//idi is a path to IM identify command.
func NewProcessor(im string, idi string) (*ImageMagickProcessor, error) {
	if len(im) == 0 {
		log.Fatal("Command convert should be set by -imConvert flag")
		return nil, errors.New("Path to imagemagick convert executable must be provided")
	}
	if len(idi) == 0 {
		log.Fatal("Command identify should be set by -imIdentify flag")
		return nil, errors.New("Path to imagemagick identify executable must be provided")
	}

	_, err := exec.LookPath(im)
	if err != nil {
		return nil, err
	}
	_, err = exec.LookPath(idi)
	if err != nil {
		return nil, err
	}

	return &ImageMagickProcessor{
		convertCmd:  im,
		identifyCmd: idi,
	}, nil
}

// Resize image to the given size preserving aspect ratio. No cropping applying.
func (p *ImageMagickProcessor) Resize(data []byte, size string) ([]byte, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size)
	args = append(args, convertOpts...)
	args = append(args, getOutputFormat(imgInfo)) //Output

	return p.execImagemagick(bytes.NewReader(data), args)
}

// Resize input image to exact size with cropping everything that out of the bounds.
// Size must specified in format WIDTHxHEIGHT. Both dimensions must be included.
func (p *ImageMagickProcessor) FitToSize(data []byte, size string) ([]byte, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size+"^")
	args = append(args, convertOpts...)
	args = append(args, cutToFitOpts...)
	args = append(args, "-extent", size)
	args = append(args, getOutputFormat(imgInfo)) //Output

	return p.execImagemagick(bytes.NewReader(data), args)
}

func (p *ImageMagickProcessor) Optimise(data []byte) ([]byte, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	quality := 82
	if imgInfo.quality > 0 && imgInfo.quality < quality {
		quality = imgInfo.quality
	}

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-quality", strconv.Itoa(quality))
	args = append(args, convertOpts...)
	args = append(args, getOutputFormat(imgInfo)) //Output

	result, err := p.execImagemagick(bytes.NewReader(data), args)
	if err != nil {
		return nil, err
	}

	if len(result) > len(data) {
		log.Printf("WARNING: Optimised size [%d] is more than original [%d], fallback to original", len(result), len(data))
		result = data
	}

	return result, nil
}

func (p *ImageMagickProcessor) execImagemagick(in *bytes.Reader, args []string) ([]byte, error) {
	var out, cmderr bytes.Buffer
	cmd := exec.Command(p.convertCmd)

	cmd.Args = append(cmd.Args, args...)

	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = &cmderr

	if Debug {
		log.Printf("Running resize command, args '%v'\n", cmd.Args)
	}
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing convert command: %s\n", err.Error())
		log.Printf("ERROR: %s\n", cmderr.String())
		return nil, err
	}

	return out.Bytes(), nil
}

func (p *ImageMagickProcessor) loadImageInfo(in *bytes.Reader) (*imageInfo, error) {
	var out, cmderr bytes.Buffer
	cmd := exec.Command(p.identifyCmd)
	cmd.Args = append(cmd.Args, "-format", "%m %Q", "-")

	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = &cmderr

	if Debug {
		log.Printf("Running identify command, args '%v'\n", cmd.Args)
	}
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing identify command: %s\n", err.Error())
		log.Printf("ERROR: %s\n", cmderr.String())
		return nil, err
	}

	imageInfo := &imageInfo{}
	fmt.Sscanf(out.String(), "%s %d", &imageInfo.format, &imageInfo.quality)

	return imageInfo, nil
}

func getOutputFormat(inf *imageInfo) string {
	output := "-"
	if inf.format == "PNG" {
		output = "PNG8:-"
	}

	return output
}
