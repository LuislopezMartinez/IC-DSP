package simpleui

import "github.com/LuislopezMartinez/IC-DSP/audio-processing-core/simpleui/internal/i18n"

func translateTitle(title string) string { return i18n.Display(title) }

// SetTranslator configures presentation-text translation and optional extra
// glyphs that must be loaded into the embedded fonts. Call it before Run.
// A nil translator restores identity translation.
func SetTranslator(translator func(string) string, extraCodepoints []rune) {
	i18n.Configure(translator, extraCodepoints)
}
