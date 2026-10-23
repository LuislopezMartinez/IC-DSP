// Package i18n is the small translation adapter used internally by simpleui.
package i18n

import "sync"

var (
	mu         sync.RWMutex
	translator = func(text string) string { return text }
	codepoints []rune
)

var messages = map[string]string{
	"text.c1ddc8f1dc6a": "simpleui: text scale must be between 0.5 and 2",
	"text.79b5195d5d11": "simpleui: font size must be greater than zero",
	"text.8ef830cdc283": "simpleui: failed to load embedded font style %d",
	"text.66630e462059": "simpleui: SetMode requires a positive width and height",
	"text.f33544ed9154": "simpleui: SetInitialWindowSize requires a positive width and height",
	"text.3a66e1494342": "simpleui: SetMinimumSize requires a positive width and height",
	"text.a12bda9bec94": "simpleui: SetTargetFPS cannot be negative",
	"text.0e40579c6e6d": "simpleui: Run requires a draw function",
	"text.8f7d1a989b96": "Raylib: configuring window flags",
	"text.eb0d363a93a6": "Raylib: calling InitWindow (%dx%d)",
	"text.8aff8158085b": "Raylib: InitWindow finished; monitor=%d; screen=%dx%d",
	"text.497e49d6569f": "Raylib: creating main render texture",
	"text.d1284adc6ad0": "Raylib: render texture created; id=%d",
	"text.a75cdd388bc1": "Raylib: entering UI loop",
	"text.510403441580": "Raylib: monitors=%d; primary=%dx%d at (%d,%d)",
	"text.d448f1eda6ce": "Raylib: window placed on primary monitor; position=(%d,%d); size=%dx%d",
	"text.563076055ae9": "simpleui: ",
	"text.3c52e64cf72a": " must be called before Run",
	"text.a70ec4fb4218": "simpleui: range slider minimum gap is outside its range",
	"text.644767ec8f1f": "simpleui: slider maximum must be greater than minimum",
	"text.8ad45dc83078": "simpleui: slider step must be greater than zero",
	"text.ed4fad023ba1": "simpleui: cannot add a nil element",
	"text.bbba2a0d8209": "simpleui: duplicate element ID: ",
	"text.ac0ed6c5b521": "simpleui: dropdown selection is outside its item list",
	"text.5d6b03c55595": "simpleui: an element ID cannot be empty",
}

func Source(key string) string {
	if value, ok := messages[key]; ok {
		return value
	}
	return key
}

func Display(text string) string {
	mu.RLock()
	fn := translator
	mu.RUnlock()
	return fn(text)
}

func Codepoints() []rune {
	mu.RLock()
	defer mu.RUnlock()
	return append([]rune(nil), codepoints...)
}

func Configure(fn func(string) string, extra []rune) {
	mu.Lock()
	defer mu.Unlock()
	if fn == nil {
		translator = func(text string) string { return text }
	} else {
		translator = fn
	}
	codepoints = append(codepoints[:0], extra...)
}
