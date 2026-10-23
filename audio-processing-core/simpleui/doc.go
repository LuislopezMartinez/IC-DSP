// Package simpleui implements a scalable retained-mode user interface on top
// of raylib-go. Applications draw in a stable logical coordinate system while
// Canvas and Viewport map rendering and input to a resizable host window.
//
// The package is independent from GO-Zero. It embeds its fonts and only
// depends on raylib-go. Applications may optionally install a translation
// callback with SetTranslator before starting the UI.
package simpleui
