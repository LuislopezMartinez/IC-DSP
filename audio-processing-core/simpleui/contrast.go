package simpleui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const minimumControlContrast = 4.5

// EnsureTextContrast preserves the requested color when it is readable and
// otherwise selects the clearer of a very light or very dark label color.
// This keeps semantic button fills legible in every application theme.
func EnsureTextContrast(text, background rl.Color) rl.Color {
	if contrastRatio(text, background) >= minimumControlContrast {
		return text
	}
	light := rl.Color{R: 250, G: 252, B: 255, A: text.A}
	dark := rl.Color{R: 8, G: 20, B: 34, A: text.A}
	if contrastRatio(light, background) >= contrastRatio(dark, background) {
		return light
	}
	return dark
}

func contrastRatio(a, b rl.Color) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func relativeLuminance(color rl.Color) float64 {
	linear := func(channel uint8) float64 {
		value := float64(channel) / 255
		if value <= 0.04045 {
			return value / 12.92
		}
		return math.Pow((value+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(color.R) + 0.7152*linear(color.G) + 0.0722*linear(color.B)
}
