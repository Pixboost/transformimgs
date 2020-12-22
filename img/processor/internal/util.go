package internal

import (
	"fmt"
	"github.com/Pixboost/transformimgs/v2/img"
	"regexp"
	"strconv"
)

var sizeRegexp = regexp.MustCompile(`^(\d*)x?(\d*)$`)

func CalculateTargetSize(source *img.Info, target *img.Info, targetSize string) error {
	if source.Width <= 0 || source.Height <= 0 {
		return nil
	}

	var err error

	parsedSize := sizeRegexp.FindStringSubmatch(targetSize)
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
			//TODO
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
