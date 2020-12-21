package internal

import (
	"github.com/Pixboost/transformimgs/v2/img"
	"testing"
)

type test struct {
	sourceWidth    int
	sourceHeight   int
	targetSize     string
	expectedWidth  int
	expectedHeight int
}

func TestCalculateTargetSize(t *testing.T) {
	tests := []*test{
		{800, 600, "400", 400, 300},
		{800, 600, "400x", 400, 300},
		{800, 600, "x300", 400, 300},
		{800, 600, "400x300", 400, 300},
	}

	for idx, tt := range tests {
		target := &img.Info{}
		CalculateTargetSize(&img.Info{
			Width:  tt.sourceWidth,
			Height: tt.sourceHeight,
		}, target, tt.targetSize)

		if target.Width != tt.expectedWidth {
			t.Errorf("Test %d failed: Expected [%d] width, but got [%d]", idx, tt.expectedWidth, target.Width)
		}
		if target.Height != tt.expectedHeight {
			t.Errorf("Test %d failed: Expected [%d] height, but got [%d]", idx, tt.expectedHeight, target.Height)
		}
	}
}
