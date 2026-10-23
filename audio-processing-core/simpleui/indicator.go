package simpleui

import rl "github.com/gen2brain/raylib-go/raylib"

type Indicator struct {
	BaseElement
	active bool
	color  rl.Color
}

func NewIndicator(id string, x, y, diameter float32) *Indicator {
	return &Indicator{
		BaseElement: NewBaseElement(id, x, y, diameter, diameter),
		color:       currentTheme.Accent,
	}
}

func (indicator *Indicator) Active() bool            { return indicator.active }
func (indicator *Indicator) SetActive(active bool)   { indicator.active = active }
func (indicator *Indicator) SetColor(color rl.Color) { indicator.color = color }
func (indicator *Indicator) Update(Input) bool       { return false }

func (indicator *Indicator) Draw() {
	bounds := indicator.Bounds()
	color := currentTheme.IndicatorOff
	if indicator.active {
		color = indicator.color
	}
	center := rl.Vector2{X: bounds.X + bounds.Width*0.5, Y: bounds.Y + bounds.Height*0.5}
	rl.DrawCircleV(center, min(bounds.Width, bounds.Height)*0.5, color)
	rl.DrawCircleLines(int32(center.X), int32(center.Y), min(bounds.Width, bounds.Height)*0.5, currentTheme.Border)
}
