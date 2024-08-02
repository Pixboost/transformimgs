/*
illustration command takes an image in stdin and prints true if image is cartoon like, including
icons, logos, illustrations.

It prints "false" for banners, product images, photos.

The initial idea is from here: https://legacy.imagemagick.org/Usage/compare/#type_reallife
*/
package main

import (
	"fmt"
	"gopkg.in/gographics/imagick.v3/imagick"
	"io"
	"log"
	"math"
	"os"
	"sort"
)

type colorSlice []*imagick.PixelWand

func (c colorSlice) Len() int           { return len(c) }
func (c colorSlice) Less(i, j int) bool { return c[i].GetColorCount() < c[j].GetColorCount() }
func (c colorSlice) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func main() {
	imgData, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var (
		colors    colorSlice
		colorsCnt uint
	)

	mw := imagick.NewMagickWand()

	err = mw.ReadImageBlob(imgData)
	if err != nil {
		log.Fatal(err)
	}

	if (mw.GetImageWidth() * mw.GetImageHeight()) > 500*500 {
		aspectRatio := float32(mw.GetImageWidth()) / float32(mw.GetImageHeight())
		err = mw.ScaleImage(500, uint(500/aspectRatio))
		if err != nil {
			log.Fatal(err)
		}
	}

	colorsCnt, colors = mw.GetImageHistogram()
	if colorsCnt > 30000 {
		fmt.Print(false)
		return
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

	fmt.Print(colorsCntIn50Pct < 10 || (float32(colorsCntIn50Pct)/float32(colorsCnt)) <= 0.02)
}
