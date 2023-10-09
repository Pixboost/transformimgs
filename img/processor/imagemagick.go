package processor

import (
	"bytes"
	"fmt"
	"github.com/Pixboost/transformimgs/v8/img"
	"github.com/Pixboost/transformimgs/v8/img/processor/internal"
	"gopkg.in/gographics/imagick.v3/imagick"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"sort"
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
	// Some fields in the target info might not be filled, so you need to check on them!
	// Argument name and value should be in a separate array elements.
	GetAdditionalArgs func(op string, image []byte, source *img.Info, target *img.Info) []string
}

var beforeResizeConvertOpts = []string{
	"-auto-orient", // changing orientation before resize, so result width and height is correct
}

var convertOpts = []string{
	"-dither", "None",
	"-define", "jpeg:fancy-upsampling=off",
	"-define", "png:compression-filter=5",
	"-define", "png:compression-level=9",
	"-define", "png:compression-strategy=0",
	"-define", "png:exclude-chunk=bKGD,cHRM,EXIF,gAMA,iCCP,iTXt,sRGB,tEXt,zCCP,zTXt,date",
	"-define", "heic:speed=6",
	"-interlace", "None",
	"-colorspace", "sRGB",
	"-sampling-factor", "4:2:0",
	"+profile", "!icc,*",
}

var cutToFitOpts = []string{
	"-gravity", "center",
}

// Debug is a flag for logging.
// When true, all IM commands will be printed to stdout.
var Debug = true

const (
	MaxWebpWidth  = 16383
	MaxWebpHeight = 16383

	// MaxAVIFTargetSize is a maximum size in pixels of the result image
	// that could be converted to AVIF.
	//
	// This is mainly done because encoding to AVIF consumes a lot of memory, and CPU time
	MaxAVIFTargetSize = 2000 * 2000

	MaxJxlLossyTargetSize = 1000 * 1000

	JxlMime  = "image/jxl"
	WebpMime = "image/webp"
	AvifMime = "image/avif"
)

func init() {
	imagick.Initialize()

	mw := imagick.NewMagickWand()
	// time resource limit is static and doesn't work with long-running processes, hence disabling it
	err := mw.SetResourceLimit(imagick.RESOURCE_TIME, -1)
	if err != nil {
		img.Log.Errorf("failed to init ImageMagick, could not set resource limit: %s, exiting...", err)
		os.Exit(1)
	}
	mw.Destroy()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		imagick.Terminate()
	}()
}

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
	source, err := p.LoadImageInfo(config.Src)
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
	args = append(args, getBeforeTransformConvertFormatOptions(source, mimeType)...)
	args = append(args, beforeResizeConvertOpts...)
	args = append(args, "-resize", targetSize)
	args = append(args, getQualityOptions(source, config, mimeType)...)
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("resize", srcData, source, target)...)
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
	source, err := p.LoadImageInfo(config.Src)
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
	args = append(args, getBeforeTransformConvertFormatOptions(source, mimeType)...)
	args = append(args, beforeResizeConvertOpts...)
	args = append(args, "-resize", targetSize+"^")

	args = append(args, getQualityOptions(source, config, mimeType)...)
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("fit", srcData, source, target)...)
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
	source, err := p.LoadImageInfo(config.Src)
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
	args = append(args, getBeforeTransformConvertFormatOptions(source, mimeType)...)
	args = append(args, beforeResizeConvertOpts...)
	args = append(args, getQualityOptions(source, config, mimeType)...)
	args = append(args, p.AdditionalArgs...)
	if p.GetAdditionalArgs != nil {
		args = append(args, p.GetAdditionalArgs("optimise", srcData, source, target)...)
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

func (p *ImageMagick) LoadImageInfo(src *img.Image) (*img.Info, error) {
	var out, cmderr bytes.Buffer
	imgId := src.Id
	in := bytes.NewReader(src.Data)
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
		Size:         in.Size(),
		Illustration: false,
	}
	_, err = fmt.Sscanf(out.String(), "%s %d %t %d %d", &imageInfo.Format, &imageInfo.Quality, &imageInfo.Opaque, &imageInfo.Width, &imageInfo.Height)
	if err != nil {
		return nil, err
	}

	if imageInfo.Format == "PNG" {
		// IM outputs quality as 92 if no quality specified
		imageInfo.Quality = 100
		imageInfo.Illustration, err = p.isIllustration(src, imageInfo)
		if err != nil {
			return nil, err
		}
	}

	return imageInfo, nil
}

type colorSlice []*imagick.PixelWand

func (c colorSlice) Len() int           { return len(c) }
func (c colorSlice) Less(i, j int) bool { return c[i].GetColorCount() < c[j].GetColorCount() }
func (c colorSlice) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

// isIllustration returns true if image is cartoon like, including
// icons, logos, illustrations.
//
// It returns false for banners, product images, photos.
//
// We use this function to decide on lossy or lossless conversion for PNG when converting
// to the next generation format.
//
// The initial idea is from here: https://legacy.imagemagick.org/Usage/compare/#type_reallife
func (p *ImageMagick) isIllustration(src *img.Image, info *img.Info) (bool, error) {
	// Assume everything less than 20Kb is a logo
	if len(src.Data) < 20*1024 {
		return true, nil
	}

	// Assume everything bigger than 1Mb is a photo
	if len(src.Data) > 1024*1024 {
		return false, nil
	}

	var (
		colors    colorSlice
		colorsCnt uint
	)

	mw := imagick.NewMagickWand()

	err := mw.ReadImageBlob(src.Data)
	if err != nil {
		return false, err
	}

	if (info.Width * info.Height) > 500*500 {
		aspectRatio := float32(info.Width) / float32(info.Height)
		err = mw.ScaleImage(500, uint(500/aspectRatio))
		if err != nil {
			return false, err
		}
	}

	colorsCnt, colors = mw.GetImageHistogram()
	if colorsCnt > 30000 {
		return false, nil
	}

	sort.Sort(sort.Reverse(colors))

	var (
		colorIdx            int
		count               uint
		currColor           *imagick.PixelWand
		pixelsCount         = uint(0)
		totalPixelsCount    = float32(mw.GetImageHeight() * mw.GetImageWidth())
		tenPercent          = uint(totalPixelsCount * 0.1)
		fiftyPercent        = uint(totalPixelsCount * 0.5)
		isBackground        = false
		lastBackgroundColor *imagick.PixelWand
		colorsInBackground  = uint(0)
		pixelsInBackground  = uint(0)
	)

	for colorIdx, currColor = range colors {
		if pixelsCount > fiftyPercent {
			break
		}

		count = currColor.GetColorCount()

		switch {
		case colorIdx == 0:
			isBackground = true
			lastBackgroundColor = currColor
			pixelsInBackground += count
			colorsInBackground++
		case isBackground:
			// Comparing colors to find out if it's still background or not.
			// This logic addresses backgrounds with more than one similar color.
			alphaDiff := currColor.GetAlpha() - lastBackgroundColor.GetAlpha()
			redDiff := currColor.GetRed() - lastBackgroundColor.GetRed()
			greenDiff := currColor.GetGreen() - lastBackgroundColor.GetGreen()
			blueDiff := currColor.GetBlue() - lastBackgroundColor.GetBlue()
			distance :=
				math.Max(math.Pow(redDiff, 2), math.Pow(redDiff-alphaDiff, 2)) +
					math.Max(math.Pow(greenDiff, 2), math.Pow(greenDiff-alphaDiff, 2)) +
					math.Max(math.Pow(blueDiff, 2), math.Pow(blueDiff-alphaDiff, 2))
			if distance < 0.1 {
				lastBackgroundColor = currColor
				pixelsInBackground += count
				colorsInBackground++
			} else {
				isBackground = false
				if pixelsInBackground < tenPercent {
					pixelsCount = pixelsInBackground
					colorsInBackground = 0
					pixelsInBackground = 0
				} else {
					pixelsCount += count
					fiftyPercent = uint((totalPixelsCount - float32(pixelsInBackground)) * 0.5)
				}
			}
		default:
			pixelsCount += count
		}
	}

	colorsCntIn50Pct := uint(colorIdx) - colorsInBackground

	return colorsCntIn50Pct < 10 || (float32(colorsCntIn50Pct)/float32(colorsCnt)) <= 0.02, nil
}

func getOutputFormat(src *img.Info, target *img.Info, supportedFormats []string) (string, string) {
	webP := false
	avif := false
	jxl := false
	for _, f := range supportedFormats {
		if f == WebpMime && src.Height < MaxWebpHeight && src.Width < MaxWebpWidth {
			webP = true
		}

		targetSize := target.Width * target.Height
		if f == AvifMime && src.Format != "GIF" && targetSize < MaxAVIFTargetSize && targetSize != 0 {
			avif = true
		}

		if f == JxlMime && src.Format != "GIF" && (src.Illustration || targetSize < MaxJxlLossyTargetSize) {
			jxl = true
		}
	}

	switch {
	case (src.Illustration && jxl) || (jxl && !avif):
		return "jxl:-", JxlMime
	case avif && !src.Illustration:
		return "avif:-", AvifMime
	case webP:
		return "webp:-", WebpMime
	}

	return "-", ""
}

func getConvertFormatOptions(source *img.Info) []string {
	var opts []string
	if source.Illustration {
		opts = append(opts, "-define", "webp:lossless=true", "-quality", "100", "-define", "jxl:effort=8")
	}
	if source.Format != "GIF" {
		opts = append(opts, "-define", "webp:method=6")
	}

	return opts
}

func getBeforeTransformConvertFormatOptions(source *img.Info, outputMimeType string) []string {
	var opts []string

	if outputMimeType == "image/webp" && source.Format == "GIF" {
		opts = append(opts, "-coalesce")
	}

	return opts
}

func getQualityOptions(source *img.Info, config *img.TransformationConfig, outputMimeType string) []string {
	var quality int

	img.Log.Printf("[%s] Getting quality for the image, source quality: %d, quality: %d, output type: %s", config.Src.Id, source.Quality, config.Quality, outputMimeType)

	if source.Illustration {
		return []string{}
	}

	switch {
	case outputMimeType == AvifMime:
		switch {
		case source.Quality > 85:
			quality = 70
		case source.Quality > 75:
			quality = 60
		default:
			quality = 50
		}
	case outputMimeType == JxlMime:
		switch {
		case source.Quality > 85:
			quality = 82
		case source.Quality > 75:
			quality = 72
		default:
			quality = 62
		}
	case source.Quality == 100:
		quality = 82
	case config.Quality != img.DEFAULT:
		quality = source.Quality
	}

	if quality == 0 {
		return []string{}
	}

	// If using lossy compression, then we can go lower
	if quality != 100 {
		switch config.Quality {
		case img.LOW:
			quality -= 10
		case img.LOWER:
			quality -= 20
		}
	}

	return []string{"-quality", strconv.Itoa(quality)}
}
