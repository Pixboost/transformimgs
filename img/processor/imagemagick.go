package processor

import (
	"bytes"
	"fmt"
	"github.com/Pixboost/transformimgs/v8/img"
	"github.com/Pixboost/transformimgs/v8/img/processor/internal"
	"github.com/gographics/imagick/imagick"
	"os/exec"
	"sort"
	"strconv"
	"sync"
	"time"
)

type ImageMagick struct {
	convertCmd  string
	identifyCmd string
	// AdditionalArgs are static arguments that will be passed to ImageMagick "convert" command for all operations.
	// Argument name and value should be in separate array elements.
	AdditionalArgs []string
	// GetAdditionalArgs could return additional arguments for ImageMagick "convert" command.
	// "op" is the name of the operation: "optimise", "resize" or "fit".
	// Some of the fields in the target info might not be filled, so you need to check on them!
	// Argument name and value should be in a separate array elements.
	GetAdditionalArgs func(op string, image []byte, source *img.Info, target *img.Info) []string
}

var convertOpts = []string{
	"-dither", "None",
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

	// Images less than 20Kb are usually logos with text.
	// Webp is usually do a better job with those.
	MinAVIFSize = 20 * 1024
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

// IsIllustration return true if image is cartoon like, including
// icons, illustrations. However, banners are not included.
//
// The initial idea is from here: https://legacy.imagemagick.org/Usage/compare/#type_reallife
func (p *ImageMagick) IsIllustration(src *img.Image) (bool, error) {
	start := time.Now()

	mw := imagick.NewMagickWand()
	err := mw.ReadImageBlob(src.Data)
	if err != nil {
		return false, err
	}
	fmt.Printf("Read image: %d\n", time.Since(start).Milliseconds())

	imageWidth := int(mw.GetImageWidth())
	imageHeight := int(mw.GetImageHeight())
	colorsByName := make(map[string]int32, imageWidth*imageHeight)

	var writeWg sync.WaitGroup
	var readWg sync.WaitGroup
	var numThreads = 8
	var rowsPerThread = imageHeight / numThreads
	pixelsCh := make(chan string, 10000)

	readWg.Add(1)
	go func() {
		defer readWg.Done()

		for p := range pixelsCh {
			count, ok := colorsByName[p]

			if !ok {
				colorsByName[p] = 1
			} else {
				colorsByName[p] = count + 1
			}
		}
	}()

	for i := 0; i < numThreads; i++ {
		startRow := rowsPerThread * i
		regionHeight := rowsPerThread
		if startRow+regionHeight > imageHeight {
			regionHeight = imageHeight - startRow
		}

		fmt.Printf("Processing [%d] rows from [%d]\n", regionHeight, startRow)
		pi := mw.NewPixelRegionIterator(0, startRow, uint(imageWidth), uint(regionHeight))
		writeWg.Add(1)
		go func() {
			defer writeWg.Done()
			for y := 0; y < regionHeight; y++ {
				pixels := pi.GetNextIteratorRow()
				for x := 0; x < imageWidth; x++ {
					pixel := pixels[x]
					h, s, l := pixel.GetHSL()
					colorName := strconv.FormatFloat(h, 'f', -1, 64) + strconv.FormatFloat(s, 'f', -1, 64) + strconv.FormatFloat(l, 'f', -1, 64)
					pixelsCh <- colorName
				}
			}
		}()
	}
	writeWg.Wait()
	close(pixelsCh)
	readWg.Wait()
	fmt.Printf("Get Colors: %d\n", time.Since(start).Milliseconds())

	totalPixelsCount := imageWidth * imageHeight
	colorsCount := make([]int, len(colorsByName))
	for _, v := range colorsByName {
		colorsCount = append(colorsCount, int(v))
	}
	fmt.Printf("Build colors count: %d\n", time.Since(start).Milliseconds())

	sort.Sort(sort.Reverse(sort.IntSlice(colorsCount)))
	fmt.Printf("Sort: %d\n", time.Since(start).Milliseconds())

	var colorIdx, c int
	background := 0
	pixels := 0
	tenPercent := int(float32(totalPixelsCount) * 0.1)
	fiftyPercent := int(float32(totalPixelsCount) * 0.5)

	for colorIdx, c = range colorsCount {
		if colorIdx == 0 && c >= tenPercent {
			background = c
			fiftyPercent = int((float32(totalPixelsCount) - float32(background)) * 0.5)
			continue
		}

		if pixels > fiftyPercent {
			break
		}

		pixels += c
	}
	fmt.Printf("Colors Iteration: %d\n", time.Since(start).Milliseconds())

	fmt.Printf("[%d] of [%d] with pixels = [%d]\n", colorIdx, len(colorsCount), fiftyPercent)

	return colorIdx*500 < fiftyPercent, nil
}

func getOutputFormat(src *img.Info, target *img.Info, supportedFormats []string) (string, string) {
	webP := false
	avif := false
	for _, f := range supportedFormats {
		if f == "image/webp" && src.Height < MaxWebpHeight && src.Width < MaxWebpWidth {
			webP = true
		}

		targetSize := target.Width * target.Height
		if f == "image/avif" && src.Format != "GIF" && src.Format != "PNG" && src.Size > MinAVIFSize && targetSize < MaxAVIFTargetSize && targetSize != 0 {
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
