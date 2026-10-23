// Package simpleui provides scalable, retained-mode UI components for raylib.
package simpleui

import rl "github.com/gen2brain/raylib-go/raylib"

// ScaleMode controls how the logical design is mapped to the host window.
type ScaleMode uint8

const (
	// Fit preserves the aspect ratio and letterboxes when necessary.
	Fit ScaleMode = iota
	// Fill preserves the aspect ratio and crops the excess.
	Fill
	// Stretch uses all available space and can change the aspect ratio.
	Stretch
)

// Viewport maps a stable logical coordinate system to a resizable window.
type Viewport struct {
	LogicalWidth  float32
	LogicalHeight float32
	Mode          ScaleMode

	windowWidth  float32
	windowHeight float32
	scaleX       float32
	scaleY       float32
	offsetX      float32
	offsetY      float32
}

func NewViewport(width, height float32, mode ScaleMode) *Viewport {
	v := &Viewport{LogicalWidth: width, LogicalHeight: height, Mode: mode}
	v.Update(int(width), int(height))
	return v
}

// Update recalculates the transform after a window size change.
func (v *Viewport) Update(windowWidth, windowHeight int) {
	v.windowWidth = max(float32(windowWidth), 1)
	v.windowHeight = max(float32(windowHeight), 1)

	sx := v.windowWidth / v.LogicalWidth
	sy := v.windowHeight / v.LogicalHeight

	switch v.Mode {
	case Stretch:
		v.scaleX, v.scaleY = sx, sy
	case Fill:
		scale := max(sx, sy)
		v.scaleX, v.scaleY = scale, scale
	default:
		scale := min(sx, sy)
		v.scaleX, v.scaleY = scale, scale
	}

	v.offsetX = (v.windowWidth - v.LogicalWidth*v.scaleX) * 0.5
	v.offsetY = (v.windowHeight - v.LogicalHeight*v.scaleY) * 0.5
}

func (v *Viewport) Destination() rl.Rectangle {
	return rl.Rectangle{
		X:      v.offsetX,
		Y:      v.offsetY,
		Width:  v.LogicalWidth * v.scaleX,
		Height: v.LogicalHeight * v.scaleY,
	}
}

// ScreenToLogical converts mouse/touch coordinates into design coordinates.
func (v *Viewport) ScreenToLogical(point rl.Vector2) rl.Vector2 {
	return rl.Vector2{
		X: (point.X - v.offsetX) / v.scaleX,
		Y: (point.Y - v.offsetY) / v.scaleY,
	}
}

func (v *Viewport) MousePosition() rl.Vector2 {
	return v.ScreenToLogical(rl.GetMousePosition())
}

func (v *Viewport) ContainsScreenPoint(point rl.Vector2) bool {
	return rl.CheckCollisionPointRec(point, v.Destination())
}

func (v *Viewport) Scale() rl.Vector2 {
	return rl.Vector2{X: v.scaleX, Y: v.scaleY}
}

func (v *Viewport) Offset() rl.Vector2 {
	return rl.Vector2{X: v.offsetX, Y: v.offsetY}
}
