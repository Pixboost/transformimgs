package internal

import (
	"fmt"
	"github.com/Pixboost/transformimgs/v7/img"
	"regexp"
	"strconv"
)

var (
	resizeRegexp = regexp.MustCompile(`^(\d*)[x]?(\d*)$`)
	fitRegexp    = regexp.MustCompile(`^(\d*)x(\d*)$`)
)

func CalculateTargetSizeForFit(target *img.Info, targetSize string) error {
	parsedSize := fitRegexp.FindStringSubmatch(targetSize)
	if len(parsedSize) < 3 || len(parsedSize[1]) == 0 || len(parsedSize[2]) == 0 {
		return fmt.Errorf("expected target size in format [WIDTH]x[HEIGHT], but got [%s]", targetSize)
	}

	var err error

	target.Width, err = strconv.Atoi(parsedSize[1])
	if err != nil {
		return fmt.Errorf("expected target size in format [WIDTH]x[HEIGHT], but got [%s]", targetSize)
	}

	target.Height, err = strconv.Atoi(parsedSize[2])
	if err != nil {
		return fmt.Errorf("expected target size in format [WIDTH]x[HEIGHT], but got [%s]", targetSize)
	}

	return nil
}

func CalculateTargetSizeForResize(source *img.Info, target *img.Info, targetSize string) error {
	if source.Width <= 0 || source.Height <= 0 {
		return nil
	}

	var err error

	parsedSize := resizeRegexp.FindStringSubmatch(targetSize)
	if len(parsedSize) < 3 {
		return fmt.Errorf("expected target size in format [WIDTH]x[HEIGHT], but got [%s]", targetSize)
	}

	if len(parsedSize[1]) > 0 {
		target.Width, err = strconv.Atoi(parsedSize[1])
		if err != nil {
			return fmt.Errorf("expected target size in format [WIDTH]x[HEIGHT], but got [%s]", targetSize)
		}
	}
	// If width specified then height will follow aspect ratio
	if len(parsedSize[2]) > 0 && target.Width == 0 {
		target.Height, err = strconv.Atoi(parsedSize[2])
		if err != nil {
			return fmt.Errorf("expected target size in format [WIDTH]x[HEIGHT], but got [%s]", targetSize)
		}
	}
	aspectRatio := float32(source.Width) / float32(source.Height)
	if target.Width > 0 {
		target.Height = int(float32(target.Width) / aspectRatio)
	} else if target.Height > 0 {
		target.Width = int(float32(target.Height) * aspectRatio)
	}

	return nil
}
