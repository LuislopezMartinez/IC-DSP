package simpleui

import "testing"

func TestSafeWindowBoundsCentersAndFitsPrimaryMonitor(t *testing.T) {
	width, height, x, y := safeWindowBounds(1600, 900, 1366, 768, 0, 0)
	if width > 1366 || height > 768 || x < 0 || y < 0 {
		t.Fatalf("window outside primary monitor: %dx%d at %d,%d", width, height, x, y)
	}
	if x != (1366-width)/2 || y != (768-height)/2 {
		t.Fatalf("window not centered: %dx%d at %d,%d", width, height, x, y)
	}
}

func TestSafeWindowBoundsUsesPrimaryMonitorOrigin(t *testing.T) {
	width, height, x, y := safeWindowBounds(1200, 700, 1920, 1080, -1920, 120)
	if width != 1200 || height != 700 || x != -1560 || y != 310 {
		t.Fatalf("unexpected positioned bounds: %dx%d at %d,%d", width, height, x, y)
	}
}
