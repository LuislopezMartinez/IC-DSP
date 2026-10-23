package simpleui

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestFitViewport(t *testing.T) {
	v := NewViewport(1600, 900, Fit)
	v.Update(1920, 1200)

	destination := v.Destination()
	closeTo(t, destination.X, 0)
	closeTo(t, destination.Y, 60)
	closeTo(t, destination.Width, 1920)
	closeTo(t, destination.Height, 1080)

	logical := v.ScreenToLogical(rl.Vector2{X: 960, Y: 600})
	closeTo(t, logical.X, 800)
	closeTo(t, logical.Y, 450)
}

func TestStretchViewport(t *testing.T) {
	v := NewViewport(1600, 900, Stretch)
	v.Update(1920, 1200)

	scale := v.Scale()
	closeTo(t, scale.X, 1.2)
	closeTo(t, scale.Y, 1200.0/900.0)
	closeTo(t, v.Destination().X, 0)
	closeTo(t, v.Destination().Y, 0)
}

func TestFillViewport(t *testing.T) {
	v := NewViewport(1600, 900, Fill)
	v.Update(1920, 1200)

	destination := v.Destination()
	closeTo(t, destination.X, (1920-1600*(1200.0/900.0))/2)
	closeTo(t, destination.Y, 0)
}

func closeTo(t *testing.T, actual, expected float32) {
	t.Helper()
	if math.Abs(float64(actual-expected)) > 0.001 {
		t.Fatalf("got %.4f, want %.4f", actual, expected)
	}
}
