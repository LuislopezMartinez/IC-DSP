package simpleui

import rl "github.com/gen2brain/raylib-go/raylib"

// Canvas renders at a fixed logical resolution and presents it in a resizable window.
type Canvas struct {
	Viewport *Viewport
	target   rl.RenderTexture2D
}

func NewCanvas(width, height int32, mode ScaleMode) *Canvas {
	canvas := &Canvas{
		Viewport: NewViewport(float32(width), float32(height), mode),
		target:   rl.LoadRenderTexture(width, height),
	}
	rl.SetTextureFilter(canvas.target.Texture, runtime.canvasFilter)
	return canvas
}

func (c *Canvas) Unload() {
	rl.UnloadRenderTexture(c.target)
}

// Begin starts drawing in logical coordinates.
func (c *Canvas) Begin(background rl.Color) {
	c.Viewport.Update(rl.GetScreenWidth(), rl.GetScreenHeight())
	rl.BeginTextureMode(c.target)
	rl.ClearBackground(background)
}

// End presents the logical canvas in the host window.
func (c *Canvas) End(letterboxColor rl.Color) {
	rl.EndTextureMode()

	rl.BeginDrawing()
	rl.ClearBackground(letterboxColor)
	rl.DrawTexturePro(
		c.target.Texture,
		rl.Rectangle{X: 0, Y: 0, Width: float32(c.target.Texture.Width), Height: -float32(c.target.Texture.Height)},
		c.Viewport.Destination(),
		rl.Vector2{},
		0,
		rl.White,
	)
	rl.EndDrawing()
}
