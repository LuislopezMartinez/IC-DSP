package simpleui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestEnsureTextContrastRepairsDarkTextOnRed(t *testing.T) {
	red := rl.Color{R: 214, G: 48, B: 55, A: 255}
	darkText := rl.Color{R: 22, G: 48, B: 82, A: 255}
	result := EnsureTextContrast(darkText, red)
	if contrastRatio(result, red) < minimumControlContrast {
		t.Fatalf("contrast remains too low: %.2f", contrastRatio(result, red))
	}
	if result == darkText {
		t.Fatal("expected the unreadable requested color to be replaced")
	}
}

func TestEnsureTextContrastPreservesReadableColor(t *testing.T) {
	dark := rl.Color{R: 10, G: 14, B: 19, A: 255}
	light := rl.Color{R: 240, G: 244, B: 246, A: 255}
	if result := EnsureTextContrast(light, dark); result != light {
		t.Fatalf("readable color changed from %#v to %#v", light, result)
	}
}
