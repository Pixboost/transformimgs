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
	error          string
}

func TestCalculateTargetSize(t *testing.T) {
	tests := []*test{
		{800, 600, "400", 400, 300, ""},
		{800, 600, "400x", 400, 300, ""},
		{800, 600, "x300", 400, 300, ""},
		{800, 600, "400x300", 400, 300, ""},
		{800, 600, "400ab300", 0, 0, "expected target size in format [WIDTH]x[HEIGHT], but got [400ab300]"},
		{0, 0, "400x300", 0, 0, ""},
		{800, 600, "abc", 0, 0, "expected target size in format [WIDTH]x[HEIGHT], but got [abc]"},
		{800, 600, "abx400", 0, 0, "expected target size in format [WIDTH]x[HEIGHT], but got [abx400]"},
		{800, 600, "300xabc", 0, 0, "expected target size in format [WIDTH]x[HEIGHT], but got [300xabc]"},
	}

	for idx, tt := range tests {
		target := &img.Info{}
		err := CalculateTargetSize(&img.Info{
			Width:  tt.sourceWidth,
			Height: tt.sourceHeight,
		}, target, tt.targetSize)

		if target.Width != tt.expectedWidth {
			t.Errorf("Test %d failed: Expected [%d] width, but got [%d]", idx, tt.expectedWidth, target.Width)
		}
		if target.Height != tt.expectedHeight {
			t.Errorf("Test %d failed: Expected [%d] height, but got [%d]", idx, tt.expectedHeight, target.Height)
		}
		if len(tt.error) == 0 && err != nil {
			t.Errorf("Test %d failed: Expected no error, but got [%s]", idx, err)
		}
		if len(tt.error) > 0 && err.Error() != tt.error {
			t.Errorf("Test %d failed: mismatched errors. Expected [%s], but got [%s]", idx, tt.error, err.Error())
		}
	}
}
