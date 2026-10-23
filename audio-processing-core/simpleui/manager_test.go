package simpleui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestManagerRegistersElementByID(t *testing.T) {
	manager := NewManager()
	button := NewButton("test", 10, 20, 100, 40, "Test", 16)
	manager.Add(button)

	if manager.Get("test") != button {
		t.Fatal("manager did not return the registered element")
	}
}

func TestButtonClicksOnlyAfterPressAndReleaseInside(t *testing.T) {
	button := NewButton("test", 10, 20, 100, 40, "Test", 16)
	clicks := 0
	button.OnClick(func() { clicks++ })
	inside := rl.Vector2{X: 30, Y: 30}

	button.Update(Input{Pointer: inside, PointerInCanvas: true, Pressed: true, Down: true})
	button.Update(Input{Pointer: inside, PointerInCanvas: true, Released: true})

	if clicks != 1 {
		t.Fatalf("got %d clicks, want 1", clicks)
	}

	button.Update(Input{Pointer: rl.Vector2{}, PointerInCanvas: true, Pressed: true})
	button.Update(Input{Pointer: inside, PointerInCanvas: true, Released: true})
	if clicks != 1 {
		t.Fatalf("outside press generated a click; got %d clicks", clicks)
	}
}
