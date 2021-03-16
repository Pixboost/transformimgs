package processor

import (
	"bytes"
	"fmt"
	"github.com/Pixboost/transformimgs/v7/img"
	"github.com/Pixboost/transformimgs/v7/img/processor/internal"
	"os/exec"
	"strconv"
)

type ImageMagick struct {
	convertCmd  string
	identifyCmd string
	// AdditionalArgs are static arguments that will be passed to ImageMagick "convert" command for all operations.
	// Argument name and value should be in separate array elements.
	AdditionalArgs []string
	// GetAdditionalArgs could return additional arguments for ImageMagick "convert" command.
	// "op" is the name of the operation: "optimise", "resize" or "fit".
	// Argument name and value should be in a separate array elements.
	GetAdditionalArgs func(op string, image []byte, imageInfo *img.Info) []string
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

	// There are two aspects to this:
	// * Encoding to AVIF consumes a lot of memory
	// * On big sizes quality of Webp is better (could be a codec thing rather than a format)
	MaxAVIFTargetSize = 1000 * 1000
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
func (p *ImageMagick) Resize(config *img.TransformationConfig) (*img.Image, error) {
	srcData := config.Src.Data
	source, err := p.loadImageInfo(bytes.NewReader(srcData), config.Src.Id)
	if err != nil {
		return nil, err
	}

	resizeConfig, ok := config.Config.(*img.ResizeConfig)
	if !ok {
		return nil, fmt.Errorf("could not get resizeConfig")
	}

	targetSize := resizeConfig.Size
	target := &img.Info{
		Opaque: source.Opaque,
	}
	err = internal.CalculateTargetSizeForResize(source, target, targetSize)
	if err != nil {
		img.Log.Errorf("could not calculate target size for [%s], targetSize: [%s]\n", config.Src.Id, targetSize)
	}
	outputFormatArg, mimeType := getOutputFormat(source, target, config.SupportedFormats)

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", targetSize)
	args = append(args, getQualityOptions(source, config, mimeType)...)
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("resize", srcData, source)...)
	}
	args = append(args, convertOpts...)
	args = append(args, getConvertFormatOptions(source)...)
	args = append(args, outputFormatArg) //Output

	outputImageData, err := p.execImagemagick(bytes.NewReader(srcData), args, config.Src.Id)
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
func (p *ImageMagick) FitToSize(config *img.TransformationConfig) (*img.Image, error) {
	srcData := config.Src.Data
	source, err := p.loadImageInfo(bytes.NewReader(srcData), config.Src.Id)
	if err != nil {
		return nil, err
	}

	resizeConfig, ok := config.Config.(*img.ResizeConfig)
	if !ok {
		return nil, fmt.Errorf("could not get resizeConfig")
	}

	targetSize := resizeConfig.Size
	target := &img.Info{
		Opaque: source.Opaque,
	}
	err = internal.CalculateTargetSizeForFit(target, targetSize)
	if err != nil {
		img.Log.Errorf("could not calculate target size for [%s], targetSize: [%s]\n", config.Src.Id, targetSize)
	}
	outputFormatArg, mimeType := getOutputFormat(source, target, config.SupportedFormats)

	args := make([]string, 0)
	args = append(args, "-") //Input
	args = append(args, "-resize", targetSize+"^")

	args = append(args, getQualityOptions(source, config, mimeType)...)
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("fit", srcData, source)...)
	}
	args = append(args, convertOpts...)
	args = append(args, cutToFitOpts...)
	args = append(args, "-extent", targetSize)
	args = append(args, getConvertFormatOptions(source)...)
	args = append(args, outputFormatArg) //Output

	outputImageData, err := p.execImagemagick(bytes.NewReader(srcData), args, config.Src.Id)
	if err != nil {
		return nil, err
	}

	return &img.Image{
		Data:     outputImageData,
		MimeType: mimeType,
	}, nil
}

func (p *ImageMagick) Optimise(config *img.TransformationConfig) (*img.Image, error) {
	srcData := config.Src.Data
	source, err := p.loadImageInfo(bytes.NewReader(srcData), config.Src.Id)
	if err != nil {
		return nil, err
	}

	target := &img.Info{
		Opaque: source.Opaque,
		Width:  source.Width,
		Height: source.Height,
	}
	outputFormatArg, mimeType := getOutputFormat(source, target, config.SupportedFormats)

	args := make([]string, 0)
	args = append(args, "-") //Input

	args = append(args, getQualityOptions(source, config, mimeType)...)
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("optimise", srcData, source)...)
	}
	args = append(args, convertOpts...)
	args = append(args, getConvertFormatOptions(source)...)
	args = append(args, outputFormatArg) //Output

	result, err := p.execImagemagick(bytes.NewReader(srcData), args, config.Src.Id)
	if err != nil {
		return nil, err
	}

	if len(result) > len(srcData) {
		img.Log.Printf("[%s] WARNING: Optimised size [%d] is more than original [%d], fallback to original", config.Src.Id, len(result), len(srcData))
		result = srcData
		mimeType = ""
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

func (p *ImageMagick) loadImageInfo(in *bytes.Reader, imgId string) (*img.Info, error) {
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

	imageInfo := &img.Info{
		Size: in.Size(),
	}
	_, err = fmt.Sscanf(out.String(), "%s %d %t %d %d", &imageInfo.Format, &imageInfo.Quality, &imageInfo.Opaque, &imageInfo.Width, &imageInfo.Height)
	if err != nil {
		return nil, err
	}

	return imageInfo, nil
}

func getOutputFormat(src *img.Info, target *img.Info, supportedFormats []string) (string, string) {
	webP := false
	avif := false
	for _, f := range supportedFormats {
		if f == "image/webp" && src.Height < MaxWebpHeight && src.Width < MaxWebpWidth {
			webP = true
		}

		targetSize := target.Width * target.Height
		if f == "image/avif" && src.Format != "GIF" && src.Size > 20*1024 && targetSize < MaxAVIFTargetSize && targetSize != 0 {
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

func getConvertFormatOptions(source *img.Info) []string {
	var opts []string
	if source.Format == "PNG" {
		opts = append(opts, "-define", "webp:lossless=true")
		if source.Opaque {
			opts = append(opts, "-colors", "256")
		}

	}
	if source.Format != "GIF" {
		opts = append(opts, "-define", "webp:method=6")
	}

	return opts
}

func getQualityOptions(source *img.Info, config *img.TransformationConfig, outputMimeType string) []string {
	var quality int

	// Lossless compression for PNG -> AVIF
	if source.Format == "PNG" && outputMimeType == "image/avif" {
		quality = 100
	} else if source.Quality == 100 {
		quality = 82
	} else if outputMimeType == "image/avif" {
		if source.Quality > 85 {
			quality = 70
		} else if source.Quality > 75 {
			quality = 60
		} else {
			quality = 50
		}
	} else if config.Quality == img.LOW {
		quality = source.Quality
	}

	if quality == 0 {
		return []string{}
	}
	if quality != 100 && config.Quality == img.LOW {
		quality -= 10
	}

	return []string{"-quality", strconv.Itoa(quality)}
}
