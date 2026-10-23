package simpleui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestButtonActivationFeedbackOnlyOnCompletedClick(t *testing.T) {
	count := 0
	SetActivationFeedback(func() { count++ })
	defer SetActivationFeedback(nil)

	button := NewButton("feedback", 10, 10, 100, 30, "BUTTON", 12)
	button.OnClick(func() {})
	inside := rl.Vector2{X: 20, Y: 20}
	button.Update(Input{Pointer: inside, PointerInCanvas: true, Pressed: true, Down: true})
	button.Update(Input{Pointer: inside, PointerInCanvas: true, Released: true})
	if count != 1 {
		t.Fatalf("feedback count = %d, want 1", count)
	}

	button.Update(Input{Pointer: inside, PointerInCanvas: true, Pressed: true, Down: true})
	button.Update(Input{Pointer: rl.Vector2{X: 500, Y: 500}, PointerInCanvas: true, Released: true})
	if count != 1 {
		t.Fatalf("cancelled click emitted feedback; count = %d", count)
	}
}
