package simpleui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestSwitchTogglesOnReleaseInside(t *testing.T) {
	toggle := NewSwitch("mute", 10, 10, 180, 40, "MUTE", false, 16)
	changes := 0
	last := false
	toggle.OnChange(func(active bool) {
		changes++
		last = active
	})
	inside := rl.Vector2{X: 40, Y: 25}

	toggle.Update(Input{Pointer: inside, PointerInCanvas: true, Pressed: true, Down: true})
	toggle.Update(Input{Pointer: inside, PointerInCanvas: true, Released: true})

	if !toggle.Active() || !last || changes != 1 {
		t.Fatalf("switch state=%v callback=%v changes=%d", toggle.Active(), last, changes)
	}
}

func TestSwitchDoesNotToggleWhenReleasedOutside(t *testing.T) {
	toggle := NewSwitch("mute", 10, 10, 180, 40, "MUTE", false, 16)
	toggle.Update(Input{Pointer: rl.Vector2{X: 40, Y: 25}, PointerInCanvas: true, Pressed: true, Down: true})
	toggle.Update(Input{Pointer: rl.Vector2{X: 400, Y: 400}, PointerInCanvas: true, Released: true})
	if toggle.Active() {
		t.Fatal("switch toggled after release outside")
	}
}

func TestDisabledSwitchIgnoresInput(t *testing.T) {
	toggle := NewSwitch("mute", 10, 10, 180, 40, "MUTE", false, 16)
	toggle.SetEnabled(false)
	inside := rl.Vector2{X: 40, Y: 25}
	toggle.Update(Input{Pointer: inside, PointerInCanvas: true, Pressed: true, Down: true})
	toggle.Update(Input{Pointer: inside, PointerInCanvas: true, Released: true})
	if toggle.Active() {
		t.Fatal("disabled switch toggled")
	}
}
