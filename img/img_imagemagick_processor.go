package img

import (
	"bytes"
	"fmt"
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
	opaque bool
}

var convertOpts = []string{
	"-unsharp", "0.25x0.08+8.3+0.045",
	"-dither", "None",
	"-posterize", "136",
	"-define", "jpeg:fancy-upsampling=off",
	"-define", "png:compression-filter=5",
	"-define", "png:compression-level=9",
	"-define", "png:compression-strategy=0",
	"-define", "webp:method=6",
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
		Log.Error("Command convert should be set by -imConvert flag")
		return nil, fmt.Errorf("Path to imagemagick convert executable must be provided")
	}
	if len(idi) == 0 {
		Log.Error("Command identify should be set by -imIdentify flag")
		return nil, fmt.Errorf("Path to imagemagick identify executable must be provided")
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
func (p *ImageMagickProcessor) Resize(data []byte, size string, imgId string) ([]byte, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data), imgId)
	if err != nil {
		return nil, err
	}

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size)
	args = append(args, convertOpts...)
	args = append(args, getConvertFormatOptions(imgInfo)...)
	args = append(args, getOutputFormat(imgInfo, []string{})) //Output

	return p.execImagemagick(bytes.NewReader(data), args, imgId)
}

// Resize input image to exact size with cropping everything that out of the bounds.
// Size must specified in format WIDTHxHEIGHT. Both dimensions must be included.
func (p *ImageMagickProcessor) FitToSize(data []byte, size string, imgId string) ([]byte, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data), imgId)
	if err != nil {
		return nil, err
	}

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size+"^")
	args = append(args, convertOpts...)
	args = append(args, cutToFitOpts...)
	args = append(args, "-extent", size)
	args = append(args, getConvertFormatOptions(imgInfo)...)
	args = append(args, getOutputFormat(imgInfo, []string{})) //Output

	return p.execImagemagick(bytes.NewReader(data), args, imgId)
}

func (p *ImageMagickProcessor) Optimise(data []byte, imgId string, supportedFormats []string) ([]byte, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data), imgId)
	if err != nil {
		return nil, err
	}
	quality := -1
	//Only changing quality if it wasn't set in original image
	if imgInfo.quality == 100 {
		quality = 82
	}

	args := make([]string, 0)
	args = append(args, "-") //Input
	if quality > 0 {
		args = append(args, "-quality", strconv.Itoa(quality))
	}
	args = append(args, convertOpts...)
	args = append(args, getConvertFormatOptions(imgInfo)...)
	args = append(args, getOutputFormat(imgInfo, supportedFormats)) //Output

	result, err := p.execImagemagick(bytes.NewReader(data), args, imgId)
	if err != nil {
		return nil, err
	}

	if len(result) > len(data) {
		Log.Printf("[%s] WARNING: Optimised size [%d] is more than original [%d], fallback to original", imgId, len(result), len(data))
		result = data
	}

	return result, nil
}

func (p *ImageMagickProcessor) execImagemagick(in *bytes.Reader, args []string, imgId string) ([]byte, error) {
	var out, cmderr bytes.Buffer
	cmd := exec.Command(p.convertCmd)

	cmd.Args = append(cmd.Args, args...)

	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = &cmderr

	if Debug {
		Log.Printf("[%s] Running resize command, args '%v'\n", imgId, cmd.Args)
	}
	err := cmd.Run()
	if err != nil {
		Log.Printf("[%s] Error executing convert command: %s\n", imgId, err.Error())
		Log.Printf("[%s] ERROR: %s\n", imgId, cmderr.String())
		return nil, err
	}

	return out.Bytes(), nil
}

func (p *ImageMagickProcessor) loadImageInfo(in *bytes.Reader, imgId string) (*imageInfo, error) {
	var out, cmderr bytes.Buffer
	cmd := exec.Command(p.identifyCmd)
	cmd.Args = append(cmd.Args, "-format", "%m %Q %[opaque]", "-")

	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = &cmderr

	if Debug {
		Log.Printf("[%s] Running identify command, args '%v'\n", imgId, cmd.Args)
	}
	err := cmd.Run()
	if err != nil {
		Log.Printf("[%s] Error executing identify command: %s\n", err.Error(), imgId)
		Log.Printf("[%s] ERROR: %s\n", cmderr.String(), imgId)
		return nil, err
	}

	imageInfo := &imageInfo{}
	fmt.Sscanf(out.String(), "%s %d %t", &imageInfo.format, &imageInfo.quality, &imageInfo.opaque)

	return imageInfo, nil
}

func getOutputFormat(inf *imageInfo, supportedFormats []string) string {
	webP := false
	for _, f := range supportedFormats {
		if f == "image/webp" {
			webP = true
		}
	}

	output := "-"
	if webP {
		output = "webp:-"
	}

	return output
}

func getConvertFormatOptions(inf *imageInfo) []string {
	if inf.format == "PNG" {
		opts := []string{
			"-define", "webp:lossless=true",
		}
		if inf.opaque {
			opts = append(opts, "-colors", "256")
		}
		return opts
	}

	return []string{}
}
