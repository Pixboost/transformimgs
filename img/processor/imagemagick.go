package processor

import (
	"bytes"
	"fmt"
	"github.com/Pixboost/transformimgs/img"
	"os/exec"
	"strconv"
)

type ImageMagick struct {
	convertCmd  string
	identifyCmd string
	// AdditionalArgs are static arguments that will be passed to ImageMagick "convert" command for all operations.
	// Argument name and value should be in separate array elements.
	AdditionalArgs []string
	// GetAdditionalArgs could return additional argument to ImageMagick "convert" command.
	// "op" is the name of the operation: "optimise", "resize" or "fit".
	// Argument name and value should be in separate array elements.
	GetAdditionalArgs func(op string, image []byte, imageInfo *ImageInfo) []string
}

type ImageInfo struct {
	format  string
	quality int
	opaque  bool
	width   int
	height  int
}

var convertOpts = []string{
	"-unsharp", "0.25x0.08+8.3+0.045",
	"-dither", "None",
	"-posterize", "136",
	"-define", "jpeg:fancy-upsampling=off",
	"-define", "png:compression-filter=5",
	"-define", "png:compression-level=9",
	"-define", "png:compression-strategy=0",
	"-define", "png:exclude-chunk=bKGD,cHRM,EXIF,gAMA,iCCP,iTXt,sRGB,tEXt,zCCP,zTXt,date",
	"-define", "webp:method=6",
	"-interlace", "None",
	"-colorspace", "sRGB",
	"-sampling-factor", "4:2:0",
	"+profile", "!icc,*",
}

var cutToFitOpts = []string{
	"-gravity", "center",
}

//If set then will print all commands to stdout.
var Debug bool = true

const (
	MaxWebpWidth  = 16383
	MaxWebpHeight = 16383
)

// NewImageMagick creates a new ImageMagick processor. It does require
// ImageMagick binaries to be installed on the local machine.
//
// im is a path to ImageMagick "convert" binary.
// idi is a path to ImageMagick "identify" binary.
func NewImageMagick(im string, idi string) (*ImageMagick, error) {
	if len(im) == 0 {
		img.Log.Error("Path to \"convert\" command should be set by -imConvert flag")
		return nil, fmt.Errorf("path to imagemagick convert binary must be provided")
	}
	if len(idi) == 0 {
		img.Log.Error("Path to \"identify\" command should be set by -imIdentify flag")
		return nil, fmt.Errorf("path to imagemagick identify binary must be provided")
	}

	_, err := exec.LookPath(im)
	if err != nil {
		return nil, err
	}
	_, err = exec.LookPath(idi)
	if err != nil {
		return nil, err
	}

	return &ImageMagick{
		convertCmd:     im,
		identifyCmd:    idi,
		AdditionalArgs: []string{},
	}, nil
}

// Resize resizes an image to the given size preserving aspect ratio. No cropping applies.
//
// Format of the size argument is WIDTHxHEIGHT with any of the dimension could be dropped, e.g. 300, x200, 300x200.
func (p *ImageMagick) Resize(data []byte, size string, imgId string, supportedFormats []string) (*img.Image, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data), imgId)
	if err != nil {
		return nil, err
	}

	outputFormatArg, mimeType := getOutputFormat(imgInfo, supportedFormats)

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size)
	if imgInfo.format == "JPEG" && imgInfo.quality < 82 {
		args = append(args, "-quality", "82")
	}
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("resize", data, imgInfo)...)
	}
	args = append(args, convertOpts...)
	args = append(args, getConvertFormatOptions(imgInfo)...)
	args = append(args, outputFormatArg) //Output

	outputImageData, err := p.execImagemagick(bytes.NewReader(data), args, imgId)
	if err != nil {
		return nil, err
	}

	return &img.Image{
		Data:     outputImageData,
		MimeType: mimeType,
	}, nil
}

// FitToSize resizes input image to exact size with cropping everything that out of the bound.
// It doesn't respect the aspect ratio of the original image.
//
// Format of the size argument is WIDTHxHEIGHT, e.g. 300x200. Both dimensions must be included.
func (p *ImageMagick) FitToSize(data []byte, size string, imgId string, supportedFormats []string) (*img.Image, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data), imgId)
	if err != nil {
		return nil, err
	}

	outputFormatArg, mimeType := getOutputFormat(imgInfo, supportedFormats)

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", size+"^")
	if imgInfo.format == "JPEG" && imgInfo.quality < 82 {
		args = append(args, "-quality", "82")
	}
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("fit", data, imgInfo)...)
	}
	args = append(args, convertOpts...)
	args = append(args, cutToFitOpts...)
	args = append(args, "-extent", size)
	args = append(args, getConvertFormatOptions(imgInfo)...)
	args = append(args, outputFormatArg) //Output

	outputImageData, err := p.execImagemagick(bytes.NewReader(data), args, imgId)
	if err != nil {
		return nil, err
	}

	return &img.Image{
		Data:     outputImageData,
		MimeType: mimeType,
	}, nil
}

func (p *ImageMagick) Optimise(data []byte, imgId string, supportedFormats []string) (*img.Image, error) {
	imgInfo, err := p.loadImageInfo(bytes.NewReader(data), imgId)
	if err != nil {
		return nil, err
	}
	quality := -1
	//Only changing quality if it wasn't set in original image
	if imgInfo.quality == 100 {
		quality = 82
	}

	outputFormatArg, mimeType := getOutputFormat(imgInfo, supportedFormats)

	args := make([]string, 0)
	args = append(args, "-") //Input
	if quality > 0 {
		args = append(args, "-quality", strconv.Itoa(quality))
	}
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("optimise", data, imgInfo)...)
	}
	args = append(args, convertOpts...)
	args = append(args, getConvertFormatOptions(imgInfo)...)
	args = append(args, outputFormatArg) //Output

	result, err := p.execImagemagick(bytes.NewReader(data), args, imgId)
	if err != nil {
		return nil, err
	}

	if len(result) > len(data) {
		img.Log.Printf("[%s] WARNING: Optimised size [%d] is more than original [%d], fallback to original", imgId, len(result), len(data))
		result = data
	}

	return &img.Image{
		Data:     result,
		MimeType: mimeType,
	}, nil
}

func (p *ImageMagick) execImagemagick(in *bytes.Reader, args []string, imgId string) ([]byte, error) {
	var out, cmderr bytes.Buffer
	cmd := exec.Command(p.convertCmd)

	cmd.Args = append(cmd.Args, args...)

	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = &cmderr

	if Debug {
		img.Log.Printf("[%s] Running resize command, args '%v'\n", imgId, cmd.Args)
	}
	err := cmd.Run()
	if err != nil {
		img.Log.Printf("[%s] Error executing convert command: %s\n", imgId, err.Error())
		img.Log.Printf("[%s] ERROR: %s\n", imgId, cmderr.String())
		return nil, err
	}

	return out.Bytes(), nil
}

func (p *ImageMagick) loadImageInfo(in *bytes.Reader, imgId string) (*ImageInfo, error) {
	var out, cmderr bytes.Buffer
	cmd := exec.Command(p.identifyCmd)
	cmd.Args = append(cmd.Args, "-format", "%m %Q %[opaque] %w %h", "-")

	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = &cmderr

	if Debug {
		img.Log.Printf("[%s] Running identify command, args '%v'\n", imgId, cmd.Args)
	}
	err := cmd.Run()
	if err != nil {
		img.Log.Printf("[%s] Error executing identify command: %s\n", err.Error(), imgId)
		img.Log.Printf("[%s] ERROR: %s\n", cmderr.String(), imgId)
		return nil, err
	}

	imageInfo := &ImageInfo{}
	_, err = fmt.Sscanf(out.String(), "%s %d %t %d %d", &imageInfo.format, &imageInfo.quality, &imageInfo.opaque, &imageInfo.width, &imageInfo.height)
	if err != nil {
		return nil, err
	}

	return imageInfo, nil
}

func getOutputFormat(inf *ImageInfo, supportedFormats []string) (string, string) {
	webP := false
	avif := false
	for _, f := range supportedFormats {
		if f == "image/webp" && inf.height < MaxWebpHeight && inf.width < MaxWebpWidth {
			webP = true
		}
		// ImageMagick doesn't support encoding of alpha channel for AVIF.
		if f == "image/avif" && inf.opaque {
			avif = true
		}
	}

	if avif {
		return "avif:-", "image/avif"
	}
	if webP {
		return "webp:-", "image/webp"
	}

	return "-", ""
}

func getConvertFormatOptions(inf *ImageInfo) []string {
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
