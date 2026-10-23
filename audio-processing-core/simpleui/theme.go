package simpleui

import rl "github.com/gen2brain/raylib-go/raylib"

// Theme defines the shared visual language of SimpleUI controls.
type Theme struct {
	Text            rl.Color
	TextMuted       rl.Color
	Accent          rl.Color
	Control         rl.Color
	ControlHover    rl.Color
	ControlPressed  rl.Color
	ControlDisabled rl.Color
	Border          rl.Color
	IndicatorOff    rl.Color
	SliderTrack     rl.Color
	SliderSelection rl.Color
	SliderHandle    rl.Color
	DisabledTrack   rl.Color
	DisabledHandle  rl.Color
	DisabledPattern rl.Color
	SwitchOff       rl.Color
	SwitchOn        rl.Color
	InputBackground rl.Color
	InputSelection  rl.Color
	InputCaret      rl.Color
	PopupBackground rl.Color
	PopupHover      rl.Color
	CornerRadius    float32
	BorderWidth     float32
}

var currentTheme = Theme{
	Text:            rl.Color{R: 240, G: 244, B: 246, A: 255},
	TextMuted:       rl.Color{R: 135, G: 158, B: 168, A: 255},
	Accent:          rl.Color{R: 45, G: 205, B: 185, A: 255},
	Control:         rl.Color{R: 35, G: 50, B: 59, A: 255},
	ControlHover:    rl.Color{R: 47, G: 70, B: 80, A: 255},
	ControlPressed:  rl.Color{R: 28, G: 174, B: 160, A: 255},
	ControlDisabled: rl.Color{R: 43, G: 49, B: 53, A: 255},
	Border:          rl.Color{R: 69, G: 91, B: 101, A: 255},
	IndicatorOff:    rl.Color{R: 72, G: 82, B: 87, A: 255},
	SliderTrack:     rl.Color{R: 55, G: 68, B: 75, A: 255},
	SliderSelection: rl.Color{R: 45, G: 205, B: 185, A: 255},
	SliderHandle:    rl.Color{R: 225, G: 235, B: 238, A: 255},
	DisabledTrack:   rl.Color{R: 55, G: 61, B: 66, A: 255},
	DisabledHandle:  rl.Color{R: 105, G: 111, B: 116, A: 255},
	DisabledPattern: rl.Color{R: 88, G: 94, B: 99, A: 180},
	SwitchOff:       rl.Color{R: 70, G: 80, B: 86, A: 255},
	SwitchOn:        rl.Color{R: 45, G: 205, B: 185, A: 255},
	InputBackground: rl.Color{R: 20, G: 31, B: 38, A: 255},
	InputSelection:  rl.Color{R: 35, G: 125, B: 145, A: 190},
	InputCaret:      rl.Color{R: 240, G: 244, B: 246, A: 255},
	PopupBackground: rl.Color{R: 17, G: 27, B: 34, A: 255},
	PopupHover:      rl.Color{R: 38, G: 68, B: 76, A: 255},
	CornerRadius:    0.12,
	BorderWidth:     1,
}

var colorTransform = func(color rl.Color) rl.Color { return color }

func SetTheme(theme Theme) { currentTheme = theme }
func CurrentTheme() Theme  { return currentTheme }
func SetColorTransform(transform func(rl.Color) rl.Color) {
	if transform == nil {
		colorTransform = func(color rl.Color) rl.Color { return color }
		return
	}
	colorTransform = transform
}
func themedColor(color rl.Color) rl.Color { return colorTransform(color) }
